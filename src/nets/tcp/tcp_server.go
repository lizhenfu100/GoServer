package tcp

import (
	"common"
	"common/std"
	"conf"
	"encoding/binary"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"io"
	"net"
	"netConfig/meta"
	"nets/rpc"
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

func NewServer(port uint16, maxconn int32, block bool) { //"ip:port"，ip可缺省
	if conf.TestFlag_CalcQPS {
		qps.Watch()
	}
	_svr.MaxConnNum = maxconn
	_svr.closer.L = &_svr.Mutex
	if l, e := net.Listen("tcp", fmt.Sprintf(":%d", port)); e == nil {
		if _svr.listener = l; block {
			_svr.run()
		} else {
			go _svr.run()
		}
	} else {
		panic("NewTcpServer: %s" + e.Error())
	}
}
func CloseServer() { _svr.Close() }

func (self *TCPServer) run() {
	var delay time.Duration
	for {
		if conn, e := self.listener.Accept(); e == nil {
			delay = 0
			go self._HandleAcceptConn(conn)
		} else if ne, ok := e.(net.Error); ok && ne.Temporary() {
			if delay == 0 {
				delay = 5 * time.Millisecond
			} else {
				delay *= 2
			}
			if delay > time.Second {
				delay = time.Second
			}
			gamelog.Error("accept error: %s retrying in %d", e.Error(), delay/time.Millisecond)
			time.Sleep(delay)
		} else {
			gamelog.Error("accept closed: " + e.Error())
			break
		}
	}
	self.Lock()
	self.listener = nil
	self.Unlock()
	self.closer.Signal()
}
func (self *TCPServer) _HandleAcceptConn(conn net.Conn) {
	buf := make([]byte, 4)                                //收第一个包，客户端上报connId
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) //首次读，5秒超时防连接占用攻击；client无需超时限制
	if _, err := io.ReadFull(conn, buf); err != nil {
		conn.Close()
		gamelog.Error("RecvFirstMsg: %s", err.Error())
		return
	}
	conn.SetReadDeadline(time.Time{}) //后续无超时限制
	if oldId := binary.LittleEndian.Uint32(buf); oldId > 0 {
		self._ResetOldConn(conn, oldId)
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
	tcpConn := newTCPConn(conn)
	self.connmap.Store(connId, tcpConn)
	atomic.AddInt32(&self.connCnt, 1)
	gamelog.Debug("AddConn(%d)", connId)

	tcpConn.onDisConnect = func() { //Notice:回调函数须线程安全
		self.connmap.Delete(connId)
		atomic.AddInt32(&self.connCnt, -1)
		gamelog.Debug("DelConnId(%d) %v", connId, tcpConn.GetUser())
		if rpc.G_HandleFunc[enum.Rpc_net_error] != nil { //先通知逻辑线程，连接断开
			msg := common.NewNetPackCap(16)
			msg.SetMsgId(enum.Rpc_net_error)
			rpc.G_RpcQueue.Insert(tcpConn, msg)
		}
		if _, ok := tcpConn.GetUser().(*meta.Meta); ok { //再注销
			msg := common.NewNetPackCap(16)
			msg.SetMsgId(enum.Rpc_unregist)
			rpc.G_RpcQueue.Insert(tcpConn, msg)
		}
	}
	//通知client，连接被接收，下发connId、密钥等
	acceptMsg := common.NewNetPackCap(16)
	acceptMsg.SetMsgId(enum.Rpc_svr_accept)
	acceptMsg.WriteUInt32(connId)
	tcpConn.WriteMsg(acceptMsg)
	acceptMsg.Free()
	self._RunConn(tcpConn, connId)
}
func (self *TCPServer) _ResetOldConn(newconn net.Conn, oldId uint32) {
	if v, ok := self.connmap.Load(oldId); ok && v.(*TCPConn).IsClose() {
		gamelog.Debug("ResetOldConn(%d)", oldId)
		oldconn := v.(*TCPConn)
		oldconn.resetConn(newconn)
		self._RunConn(oldconn, oldId)
	} else {
		self._AddNewConn(newconn)
	}
}
func (self *TCPServer) _RunConn(conn *TCPConn, connId uint32) {
	self.wgConns.Add(1)
	go conn.writeLoop()
	conn.readLoop() //block read io，保证消息响应、连接回收，在同一线程处理
	if _, ok := conn.GetUser().(*meta.Meta); ok {
		conn.onDisConnect() //节点断开，须立即注销
	} else if conn.delayDel == nil {
		conn.delayDel = time.AfterFunc(Delay_Delete_Conn, func() {
			if conn.IsClose() {
				conn.onDisConnect()
			}
		})
	} else {
		conn.delayDel.Reset(Delay_Delete_Conn)
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

// ------------------------------------------------------------
// 模块注册
var g_reg_conn_map sync.Map //<{module,svrId}, *TCPConn>

func _Rpc_regist(req, _ *common.NetPack, conn common.Conn) {
	pMeta := new(meta.Meta)
	pMeta.BufToData(req)
	if p := meta.GetMeta(pMeta.Module, pMeta.SvrID); p != nil {
		if p.IP != pMeta.IP || p.OutIP != pMeta.OutIP {
			conn.Close() //防止配置错误，出现外网节点顶替
			gamelog.Warn("Regist repeat: %v", pMeta)
			return
		}
	}
	conn.SetUser(pMeta)
	meta.AddMeta(pMeta)
	g_reg_conn_map.Store(std.KeyPair{pMeta.Module, pMeta.SvrID}, conn)
	gamelog.Debug("Regist: %v", pMeta)
}
func _Rpc_unregist(req, _ *common.NetPack, conn common.Conn) {
	if pMeta, ok := conn.GetUser().(*meta.Meta); ok {
		if pConn := FindRegModule(pMeta.Module, pMeta.SvrID); pConn == nil || pConn.IsClose() {
			gamelog.Debug("UnRegist: %v", pMeta)
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
