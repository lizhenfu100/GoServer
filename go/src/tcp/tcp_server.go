package tcp

import (
	"common"
	"gamelog"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	listener        net.Listener
	mutexConns      sync.Mutex
	connset         map[net.Conn]bool
	wgLn            sync.WaitGroup
	wgConns         sync.WaitGroup
}

func NewTcpServer(addr string, maxconn int) {
	svr := new(TCPServer)
	svr.Addr = addr //"ip:port"，ip可缺省
	svr.MaxConnNum = maxconn
	svr.PendingWriteNum = 32
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
	if self.PendingWriteNum <= 0 {
		self.PendingWriteNum = 32
		gamelog.Info("Invalid PendingWriteNum, reset to %d", self.PendingWriteNum)
	}
	self.listener = ln
	self.connset = make(map[net.Conn]bool)
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
		connNum := len(self.connset)
		if connNum >= self.MaxConnNum {
			conn.Close()
			gamelog.Error("too many connections(%d/%d)", connNum, self.MaxConnNum)
			continue
		}

		self.mutexConns.Lock()
		self.connset[conn] = true
		self.mutexConns.Unlock()
		self.wgConns.Add(1)
		gamelog.Info("Connect From: %s,  ConnNum: %d", conn.RemoteAddr().String(), connNum+1)
		tcpConn := newTCPConn(conn, self.PendingWriteNum,
			func(this *TCPConn) {
				// 清理tcp_server相关数据
				self.mutexConns.Lock()
				delete(self.connset, this.conn)
				self.mutexConns.Unlock()
				gamelog.Info("Connect Endded:UserPtr:%v, ConnNum is:%d", this.UserPtr, len(self.connset))
				self.wgConns.Done()
			})
		go tcpConn.readRoutine()
		go tcpConn.writeRoutine()
	}
}
func (self *TCPServer) Close() {
	self.listener.Close()
	self.wgLn.Wait()

	self.mutexConns.Lock()
	for conn := range self.connset {
		conn.Close()
	}
	self.connset = nil
	self.mutexConns.Unlock()

	self.wgConns.Wait()
	gamelog.Info("server been closed!!")
}

//////////////////////////////////////////////////////////////////////
//! 模块注册
var g_reg_conn_map = make(map[common.KeyPair]*TCPConn)

func DoRegistToSvr(conn *TCPConn, data *common.NetPack) {
	module := data.ReadString()
	id := data.ReadInt()
	g_reg_conn_map[common.KeyPair{module, id}] = conn
}
func FindRegModuleConn(module string, id int) *TCPConn {
	if v, ok := g_reg_conn_map[common.KeyPair{module, id}]; ok {
		return v
	}
	gamelog.Error("FindRegModuleConn nil : (%s,%d)-%v", module, id, g_reg_conn_map)
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
