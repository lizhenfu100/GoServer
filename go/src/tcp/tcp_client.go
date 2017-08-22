package tcp

import (
	"common"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type TCPClient struct {
	addr      string
	connId    uint32
	TcpConn   *TCPConn
	OnConnect func(*TCPConn)
}

func (self *TCPClient) ConnectToSvr(addr, srcModule string, srcID int) {
	self.addr = addr
	go self.connectRoutine(srcModule, srcID) //会断线后自动重连
}
func (self *TCPClient) connectRoutine(srcModule string, srcID int) {
	regMsg := common.NewNetPackCap(32)
	regMsg.SetOpCode(G_MsgId_Regist)
	regMsg.WriteString(srcModule)
	regMsg.WriteInt(srcID)
	for {
		if self.connect() {
			if self.OnConnect != nil {
				self.OnConnect(self.TcpConn)
			}
			go self.TcpConn.readRoutine()
			//Notice: 这里不能用CallRpc，非线程安全的
			firstMsg := make([]byte, 2+4)
			binary.LittleEndian.PutUint32(firstMsg[2:], self.connId) //上报connId
			self.TcpConn.WriteBuf(firstMsg)
			self.TcpConn.WriteMsg(regMsg)
			self.TcpConn.writeRoutine() //goroutine会阻塞在这里
		}
		time.Sleep(3 * time.Second)
	}
}
func (self *TCPClient) connect() bool {
	conn, err := net.Dial("tcp", self.addr)
	if err != nil {
		fmt.Printf("connect to %s error :%s \n", self.addr, err.Error())
		return false
	}
	if conn == nil {
		return false
	}
	if self.TcpConn == nil {
		self.TcpConn = newTCPConn(conn, nil)
		self.TcpConn.UserPtr = self
	} else {
		//断线重连的新连接标记得重置，否则tcpConn.readRoutine.readLoop会直接break
		self.TcpConn.ResetConn(conn)
	}
	return true
}
func (self *TCPClient) Close() {
	self.TcpConn.Close()
	self.TcpConn = nil
}
func OnSvrAcceptConn(req, ack *common.NetPack, conn *TCPConn) {
	self := conn.UserPtr.(*TCPClient)
	self.connId = req.ReadUInt32()
}
