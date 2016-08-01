package tcp

import (
	"gamelog"
	"net"
	"sync"
	"time"
)

type ConnSet map[net.Conn]bool
type TCPServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	listener        net.Listener
	mutexConns      sync.Mutex
	connset         ConnSet
	wgLn            sync.WaitGroup
	wgConns         sync.WaitGroup
}

type MsgHanler func(pTcpConn *TCPConn, pdata []byte)

var G_HandlerMap map[int16]func(pTcpConn *TCPConn, pdata []byte)

func HandleFunc(msgid int16, mh MsgHanler) {
	if G_HandlerMap == nil {
		G_HandlerMap = make(map[int16]func(pTcpConn *TCPConn, pdata []byte), 100)
	}
	G_HandlerMap[msgid] = mh
}

func NewServer(addr string, maxconn int) {
	svr := new(TCPServer)
	svr.Addr = addr //"ip:port"，ip可缺省
	svr.MaxConnNum = maxconn
	svr.PendingWriteNum = 32
	svr.init()
	svr.run()
	svr.close()
}
func (server *TCPServer) CloseRun() {
	server.close()
}

func (server *TCPServer) init() bool {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		gamelog.Error("TCPServer Init failed  error :%s", err.Error())
		return false
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 5000
		gamelog.Info("Invalid MaxConnNum, reset to %d", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 32
		gamelog.Info("Invalid PendingWriteNum, reset to %d", server.PendingWriteNum)
	}

	server.listener = ln
	server.connset = make(ConnSet)

	return true
}
func (server *TCPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()
	var tempDelay time.Duration
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				gamelog.Error("accept error: %s retrying in %d", err.Error(), tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			gamelog.Error("accept error: %s", err.Error())
			return
		}
		tempDelay = 0
		connNum := len(server.connset)
		server.mutexConns.Lock()
		if connNum >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			gamelog.Error("too many connections")
			continue
		}

		server.connset[conn] = true
		server.mutexConns.Unlock()
		server.wgConns.Add(1)
		gamelog.Info("Connect From: %s,  ConnNum: %d", conn.RemoteAddr().String(), connNum+1)
		tcpConn := newTCPConn(conn, server.PendingWriteNum)
		tcpConn.onReadRoutineEnd = func() {
			// 清理tcp_server相关数据
			server.mutexConns.Lock()
			delete(server.connset, tcpConn.conn)
			connNum := len(server.connset)
			gamelog.Info("Connect Endded:Data:%v, ConnNum is:%d", tcpConn.Data, connNum)
			server.mutexConns.Unlock()
			server.wgConns.Done()
		}
		go tcpConn.readRoutine()
		go tcpConn.writeRoutine()
	}
}
func (server *TCPServer) close() {
	server.listener.Close()
	server.wgLn.Wait()

	server.mutexConns.Lock()
	for conn := range server.connset {
		conn.Close()
	}

	server.connset = nil
	server.mutexConns.Unlock()
	server.wgConns.Wait()
	gamelog.Info("server been closed!!")
}
