package tcp

import (
	"common"
	"encoding/binary"
	"gamelog"
	"generate_out/rpc/enum"
	"io"
	"net"
	"netConfig/meta"
	"sync"
	"sync/atomic"
	"time"
)

type TCPServer struct {
	MaxConnNum int32
	connCnt    int32
	autoConnId uint32
	connmap    sync.Map
	listener   net.Listener
	wgLn       sync.WaitGroup
	wgConns    sync.WaitGroup
}

func NewTcpServer(addr string, maxconn int) { //"ip:port"，ip可缺省
	var err error
	svr := TCPServer{MaxConnNum: int32(maxconn)}
	if svr.listener, err = net.Listen("tcp", addr); err == nil {
		svr.run()
		svr.Close()
	} else {
		panic("NewTcpServer failed :%s" + err.Error())
	}
}
func (self *TCPServer) run() {
	self.wgLn.Add(1)
	defer self.wgLn.Done()
	var tempDelay time.Duration
	for {
		conn, err := self.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if tempDelay > time.Second {
					tempDelay = time.Second
				}
				gamelog.Error("accept error: %s retrying in %d", err.Error(), tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			gamelog.Error("accept error: %s", err.Error())
			return
		}
		tempDelay = 0
		go self._HandleAcceptConn(conn)
	}
}
func (self *TCPServer) _HandleAcceptConn(conn net.Conn) {
	buf := make([]byte, 2+4)                              //Notice：收第一个包，客户端上报connId
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) //首次读，5秒超时防连接占用攻击；client无需超时限制
	if _, err := io.ReadFull(conn, buf); err != nil {
		conn.Close()
		gamelog.Error("RecvFirstPack: %s", err.Error())
		return
	}
	connId := binary.LittleEndian.Uint32(buf[2:])
	gamelog.Debug("_HandleAcceptConn: %d", connId)

	conn.SetReadDeadline(time.Time{}) //后续无超时限制

	if connId > 0 {
		self._ResetOldConn(conn, connId)
	} else {
		self._AddNewConn(conn)
	}
}
func (self *TCPServer) _AddNewConn(conn net.Conn) {
	if self.connCnt >= self.MaxConnNum {
		conn.Close()
		gamelog.Error("too many connections")
		return
	}

	connId := atomic.AddUint32(&self.autoConnId, 1)

	self.wgConns.Add(1)
	tcpConn := newTCPConn(conn)
	self.connmap.Store(connId, tcpConn)
	atomic.AddInt32(&self.connCnt, 1)
	gamelog.Debug("Connect From: %s, AddConnId: %d", conn.RemoteAddr().String(), connId)

	//延时删除，以待断线重连
	tcpConn.delayDel = time.AfterFunc(Delay_Delete_Conn, func() { //Notice:回调函数须线程安全
		if tcpConn.IsClose() {
			self.connmap.Delete(connId)
			atomic.AddInt32(&self.connCnt, -1)
			gamelog.Debug("Disconnect: DelConnId: %d %v", connId, tcpConn.UserPtr)

			//通知逻辑线程，连接断开
			packet := common.NewNetPackCap(16)
			packet.SetOpCode(enum.Rpc_net_error)
			G_RpcQueue.Insert(tcpConn, packet)

			//连接的后台节点断开，注销之
			if _, ok := tcpConn.UserPtr.(*meta.Meta); ok {
				packet := common.NewNetPackCap(16)
				packet.SetOpCode(enum.Rpc_unregist)
				G_RpcQueue.Insert(tcpConn, packet)
			}
		}
	})
	tcpConn.delayDel.Stop()

	//通知client，连接被接收，下发connId、密钥等
	acceptMsg := common.NewNetPackCap(32)
	acceptMsg.SetOpCode(enum.Rpc_svr_accept)
	acceptMsg.WriteUInt32(connId)
	tcpConn.WriteMsg(acceptMsg)
	acceptMsg.Free()
	self._RunConn(tcpConn, connId)
}
func (self *TCPServer) _ResetOldConn(newconn net.Conn, oldId uint32) {
	if v, ok := self.connmap.Load(oldId); ok {
		oldconn := v.(*TCPConn)
		if oldconn.IsClose() {
			gamelog.Debug("_ResetOldConn isClose: %d", oldId)
			oldconn.resetConn(newconn)
			self._RunConn(oldconn, oldId)
		} else {
			gamelog.Debug("_ResetOldConn isOpen: %d", oldId)
			// newconn.Close()
			self._AddNewConn(newconn)
		}
	} else { //服务器重启
		gamelog.Debug("_ResetOldConn to _AddNewConn: %d", oldId)
		self._AddNewConn(newconn)
	}
}
func (self *TCPServer) _RunConn(conn *TCPConn, connId uint32) {
	go conn.readRoutine()
	conn.writeRoutine() //block

	//【迁移至：延时删除】
	//self.connmap.Delete(connId)
	//atomic.AddInt32(&self.connCnt, -1)
	//gamelog.Debug("Disconnect: DelConnId: %d", connId)
	conn.delayDel.Reset(Delay_Delete_Conn)

	self.wgConns.Done()
}

func (self *TCPServer) Close() {
	self.listener.Close()
	self.wgLn.Wait()

	self.connmap.Range(func(k, v interface{}) bool {
		v.(*TCPConn).Close()
		return true
	})

	self.wgConns.Wait()
	gamelog.Debug("server been closed!!")
}

// ------------------------------------------------------------
// 模块注册
var g_reg_conn_map sync.Map

func _Rpc_regist(req, ack *common.NetPack, conn *TCPConn) {
	pMeta := new(meta.Meta)
	pMeta.BufToData(req)
	conn.UserPtr = pMeta

	g_reg_conn_map.Store(common.KeyPair{pMeta.Module, pMeta.SvrID}, conn)
	meta.AddMeta(pMeta)
	gamelog.Info("RegistToSvr: {%s %d}", pMeta.Module, pMeta.SvrID)
}
func _Rpc_unregist(req, ack *common.NetPack, conn *TCPConn) {
	if pMeta, ok := conn.UserPtr.(*meta.Meta); ok {
		if pConn := FindRegModule(pMeta.Module, pMeta.SvrID); pConn != nil && pConn.IsClose() {
			gamelog.Info("UnRegist Svr: {%s %d}", pMeta.Module, pMeta.SvrID)
			meta.DelMeta(pMeta.Module, pMeta.SvrID)
			g_reg_conn_map.Delete(common.KeyPair{pMeta.Module, pMeta.SvrID})
		}
	}
}

func FindRegModule(module string, id int) *TCPConn {
	if v, ok := g_reg_conn_map.Load(common.KeyPair{module, id}); ok {
		return v.(*TCPConn)
	}
	gamelog.Error("FindRegModuleConn nil : (%s,%d)", module, id)
	return nil
}

func GetRegModuleIDs(module string) (ret []int) {
	g_reg_conn_map.Range(func(k, v interface{}) bool {
		if k.(common.KeyPair).Name == module {
			ret = append(ret, k.(common.KeyPair).ID)
		}
		return true
	})
	return
}
func ForeachRegModule(f func(v *TCPConn)) {
	g_reg_conn_map.Range(func(k, v interface{}) bool {
		f(v.(*TCPConn))
		return true
	})
}
