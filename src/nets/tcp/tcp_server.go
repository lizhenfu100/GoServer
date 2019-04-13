package tcp

import (
	"common"
	"common/std"
	"common/timer"
	"conf"
	"encoding/binary"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"io"
	"net"
	"netConfig/meta"
	"svr_client/test/qps"
	"sync"
	"sync/atomic"
	"time"
)

type TCPServer struct {
	sync.Mutex
	closer     sync.Cond
	listener   net.Listener
	MaxConnNum int32
	connCnt    int32
	autoConnId uint32
	connmap    sync.Map //<connId, *TCPConn>
	wgConns    sync.WaitGroup
}

var _svr TCPServer

func NewTcpServer(port uint16, maxconn int32) { //"ip:port"，ip可缺省
	if conf.TestFlag_CalcQPS {
		go qps.WatchLoop()
	}
	_svr.MaxConnNum = maxconn
	_svr.closer.L = &_svr.Mutex
	if l, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
		_svr.SetListener(l)
		_svr.run()
	} else {
		panic("NewTcpServer failed :%s" + err.Error())
	}
}
func CloseServer() { _svr.Close() }

func (self *TCPServer) run() {
	var tempDelay time.Duration
	for {
		if conn, err := self.listener.Accept(); err == nil {
			tempDelay = 0
			go self._HandleAcceptConn(conn)
		} else {
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
			gamelog.Error("accept closed: " + err.Error())
			break
		}
	}
	self.SetListener(nil)
	self.closer.Signal()
}
func (self *TCPServer) _HandleAcceptConn(conn net.Conn) {
	buf := make([]byte, 4)                                //收第一个包，客户端上报connId
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) //首次读，5秒超时防连接占用攻击；client无需超时限制
	if _, err := io.ReadFull(conn, buf); err != nil {
		conn.Close()
		gamelog.Error("(%p)RecvFirstPack: %s", conn, err.Error())
		return
	}
	connId := binary.LittleEndian.Uint32(buf)
	gamelog.Debug("_HandleAcceptConn: %d", connId)
	//FIXME：IP黑/白名单
	conn.SetReadDeadline(time.Time{}) //后续无超时限制

	if connId > 0 {
		self._ResetOldConn(conn, connId)
	} else {
		self._AddNewConn(conn)
	}
}
func (self *TCPServer) _AddNewConn(conn net.Conn) {
	if atomic.LoadInt32(&self.connCnt) >= self.MaxConnNum {
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

	tcpConn.onDisConnect = func(conn *TCPConn) { //Notice:回调函数须线程安全
		self.connmap.Delete(connId)
		atomic.AddInt32(&self.connCnt, -1)
		gamelog.Debug("DelConnId: %d %v", connId, conn.UserPtr)

		//通知逻辑线程，连接断开
		packet := common.NewNetPackCap(16)
		packet.SetOpCode(enum.Rpc_net_error)
		G_RpcQueue.Insert(conn, packet)

		//连接的后台节点断开，注销之
		if _, ok := conn.UserPtr.(*meta.Meta); ok {
			packet := common.NewNetPackCap(16)
			packet.SetOpCode(enum.Rpc_unregist)
			G_RpcQueue.Insert(conn, packet)
		}
	}

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

	//针对玩家链接，延时删除，以待断线重连
	if _, ok := conn.UserPtr.(*meta.Meta); ok {
		conn.onDisConnect(conn)
	} else {
		//if conn.delayDel == nil {
		//	conn.delayDel = time.AfterFunc(Delay_Delete_Conn, func() {
		//		if tcpConn.IsClose() {
		//			tcpConn.onDisConnect()
		//		}
		//	})
		//} else {
		//	conn.delayDel.Reset(Delay_Delete_Conn)
		//}
		conn.delayDel = timer.G_TimerMgr.AddTimerSec(func() {
			if conn.IsClose() {
				conn.onDisConnect(conn)
			}
		}, Delay_Delete_Conn, 0, 0)
	}

	self.wgConns.Done()
}

func (self *TCPServer) Close() {
	self.Lock()
	for self.listener != nil {
		self.listener.Close()
		self.closer.Wait()
	}
	self.Unlock()

	self.connmap.Range(func(k, v interface{}) bool {
		v.(*TCPConn).Close()
		return true
	})
	self.wgConns.Wait()
	gamelog.Info("server been closed!!")
}
func (self *TCPServer) SetListener(l net.Listener) {
	self.Lock()
	self.listener = l
	self.Unlock()
}

// ------------------------------------------------------------
// 模块注册
var g_reg_conn_map sync.Map //<{module,svrId}, *TCPConn>

func _Rpc_regist(req, _ *common.NetPack, conn *TCPConn) {
	pMeta := new(meta.Meta)
	pMeta.BufToData(req)
	if p := meta.GetMeta(pMeta.Module, pMeta.SvrID); p != nil {
		if p.IP != pMeta.IP || p.OutIP != pMeta.OutIP {
			//防止配置错误，出现外网节点顶替
			conn.Close()
			fmt.Println("Error: RegistToSvr repeat: ", pMeta)
			return
		}
	}
	conn.UserPtr = pMeta
	meta.AddMeta(pMeta)
	g_reg_conn_map.Store(std.KeyPair{pMeta.Module, pMeta.SvrID}, conn)
	fmt.Println("RegistToSvr: ", pMeta)
}
func _Rpc_unregist(req, _ *common.NetPack, conn *TCPConn) {
	if pMeta, ok := conn.UserPtr.(*meta.Meta); ok {
		if pConn := FindRegModule(pMeta.Module, pMeta.SvrID); pConn == nil || pConn.IsClose() {
			fmt.Println("UnRegist Svr: ", pMeta)
			meta.DelMeta(pMeta.Module, pMeta.SvrID)
			g_reg_conn_map.Delete(std.KeyPair{pMeta.Module, pMeta.SvrID})
		}
	}
}

func FindRegModule(module string, id int) *TCPConn {
	if v, ok := g_reg_conn_map.Load(std.KeyPair{module, id}); ok {
		return v.(*TCPConn)
	}
	gamelog.Debug("FindRegModule nil : (%s,%d)", module, id)
	return nil
}
func ForeachRegModule(f func(v *TCPConn)) {
	g_reg_conn_map.Range(func(k, v interface{}) bool {
		f(v.(*TCPConn))
		return true
	})
}
