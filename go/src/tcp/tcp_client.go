package tcp

import (
	"common"
	"fmt"
	"net"
	"time"
)

type TCPClient struct {
	Addr            string
	PendingWriteNum int
	TcpConn         *TCPConn
	OnConnect       func(*TCPConn)
}

func (self *TCPClient) ConnectToSvr(addr, srcModule string, srcID int) {
	self.Addr = addr
	self.PendingWriteNum = 32
	self.TcpConn = nil

	go self.connectRoutine(srcModule, srcID) //会断线后自动重连
}
func (self *TCPClient) connectRoutine(srcModule string, srcID int) {
	packet := common.NewNetPackCap(32)
	packet.SetOpCode(G_MsgId_Regist)
	packet.WriteString(srcModule)
	packet.WriteInt(srcID)
	for {
		if self.connect() {
			if self.TcpConn != nil {
				if self.OnConnect != nil {
					self.OnConnect(self.TcpConn)
				}
				go self.TcpConn.writeRoutine()
				self.TcpConn.WriteMsg(packet)
				self.TcpConn.readRoutine() //goroutine会阻塞在这里
			}
		}
		time.Sleep(3 * time.Second)
	}
}
func (self *TCPClient) connect() bool {
	conn, err := net.Dial("tcp", self.Addr)
	if err != nil {
		fmt.Printf("connect to %s error :%s \n", self.Addr, err.Error())
		return false
	}
	if conn == nil {
		return false
	}

	if self.TcpConn != nil {
		//断线重连的新连接标记得重置，否则tcpConn.readRoutine.readLoop会直接break
		self.TcpConn.ResetConn(conn)
	} else {
		self.TcpConn = newTCPConn(conn, self.PendingWriteNum, nil)
	}
	return true
}

func (self *TCPClient) Close() {
	self.TcpConn.Close()
	self.TcpConn = nil
}
