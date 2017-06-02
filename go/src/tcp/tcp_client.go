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
	firstBuf        *common.NetPack // 连接建立后的第一个包，上报connId、密钥等
}

func (self *TCPClient) ConnectToSvr(addr, srcModule string, srcID int) {
	self.Addr = addr
	self.PendingWriteNum = 32
	self.firstBuf = common.NewNetPackLen(4)

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
			self.TcpConn.WriteMsg(self.firstBuf)
			self.TcpConn.WriteMsg(regMsg)
			self.TcpConn.writeRoutine() //goroutine会阻塞在这里
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
	if self.TcpConn == nil {
		self.TcpConn = newTCPConn(conn, self.PendingWriteNum, nil)
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
func OnSvrAcceptConn(conn *TCPConn, data *common.NetPack) {
	client := conn.UserPtr.(*TCPClient)
	connId := data.ReadUInt32()
	client.firstBuf.SetPos(0, connId)
}
