package tcp

import (
	"common"
	"encoding/binary"
	"gamelog"
	"generate/rpc/enum"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type TCPServer struct {
	Addr       string
	MaxConnNum int
	autoConnId uint32
	connmap    map[uint32]*TCPConn
	listener   net.Listener
	mutexConns sync.Mutex
	wgLn       sync.WaitGroup
	wgConns    sync.WaitGroup
}

func NewTcpServer(addr string, maxconn int) {
	svr := new(TCPServer)
	svr.Addr = addr //"ip:port"，ip可缺省
	svr.MaxConnNum = maxconn
	svr.init()
	svr.run()
	svr.Close()
}
func (self *TCPServer) init() bool {
	ln, err := net.Listen("tcp", self.Addr)
	if err != nil {
		gamelog.Error("TCPServer Init failed  error :%s", err.Error())
		return false
	}
	self.listener = ln
	self.connmap = make(map[uint32]*TCPConn)
	return true
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
	buf := make([]byte, 2+4) //Notice：收第一个包，客户端上报connId
	if _, err := io.ReadFull(conn, buf); err != nil {
		conn.Close()
		gamelog.Error("RecvFirstPack: %s", err.Error())
		return
	}
	connId := binary.LittleEndian.Uint32(buf[2:])
	gamelog.Info("_HandleAcceptConn: %d", connId)

	if connId > 0 {
		self._ResetOldConn(conn, connId)
	} else {
		self._AddNewConn(conn)
	}
}
func (self *TCPServer) _AddNewConn(conn net.Conn) {
	if len(self.connmap) >= self.MaxConnNum {
		conn.Close()
		gamelog.Error("too many connections(%d/%d)", len(self.connmap), self.MaxConnNum)
		return
	}

	connId := atomic.AddUint32(&self.autoConnId, 1)

	self.wgConns.Add(1)
	tcpConn := newTCPConn(conn,
		func(this *TCPConn) {
			self.mutexConns.Lock()
			delete(self.connmap, connId)
			self.mutexConns.Unlock()
			gamelog.Info("Disconnect: UserPtr:%v, DelConnId: %d, ConnCnt: %d", this.UserPtr, connId, len(self.connmap))
			self.wgConns.Done()
		})
	self.mutexConns.Lock()
	self.connmap[connId] = tcpConn
	self.mutexConns.Unlock()
	gamelog.Info("Connect From: %s, AddConnId: %d, ConnCnt: %d", conn.RemoteAddr().String(), connId, len(self.connmap))

	go tcpConn.readRoutine()
	// 通知client，连接被接收，下发connId、密钥等
	acceptMsg := common.NewNetPackCap(32)
	acceptMsg.SetOpCode(enum.Rpc_svr_accept)
	acceptMsg.WriteUInt32(connId)
	tcpConn.WriteMsg(acceptMsg)
	tcpConn.writeRoutine()
}
func (self *TCPServer) _ResetOldConn(newconn net.Conn, oldId uint32) {
	self.mutexConns.Lock()
	oldconn, ok := self.connmap[oldId]
	self.mutexConns.Unlock()
	if oldconn != nil && ok {
		if oldconn.isClose {
			gamelog.Info("_ResetOldConn isClose: %d", oldId)
			oldconn.ResetConn(newconn)
			go oldconn.readRoutine()
			oldconn.writeRoutine()
		} else {
			gamelog.Info("_ResetOldConn isOpen: %d", oldId)
			// newconn.Close()
			self._AddNewConn(newconn)
		}
	} else { //服务器重启
		gamelog.Info("_ResetOldConn to _AddNewConn: %d", oldId)
		self._AddNewConn(newconn)
	}
}
func (self *TCPServer) Close() {
	self.listener.Close()
	self.wgLn.Wait()

	self.mutexConns.Lock()
	for _, conn := range self.connmap {
		conn.Close()
	}
	self.connmap = nil
	self.mutexConns.Unlock()

	self.wgConns.Wait()
	gamelog.Info("server been closed!!")
}

//////////////////////////////////////////////////////////////////////
//! 模块注册
var g_reg_conn_map = make(map[common.KeyPair]*TCPConn)

func DoRegistToSvr(req, ack *common.NetPack, conn *TCPConn) {
	module := req.ReadString()
	id := req.ReadInt()
	g_reg_conn_map[common.KeyPair{module, id}] = conn
	gamelog.Info("DoRegistToSvr: {%s %d}", module, id)
}
func FindRegModuleConn(module string, id int) *TCPConn {
	if v, ok := g_reg_conn_map[common.KeyPair{module, id}]; ok {
		return v
	}
	gamelog.Error("FindRegModuleConn nil : (%s,%d) => %v", module, id, g_reg_conn_map)
	return nil
}
func GetRegModuleIDs(module string) (ret []int) {
	for k, _ := range g_reg_conn_map {
		if k.Name == module {
			ret = append(ret, k.ID)
		}
	}
	return
}
