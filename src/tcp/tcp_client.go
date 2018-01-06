package tcp

import (
	"common"
	"common/net/meta"
	"encoding/binary"
	"fmt"
	"generate_out/rpc/enum"
	"net"
	"time"
)

type TCPClient struct {
	addr      string
	connId    uint32
	TcpConn   *TCPConn
	OnConnect func(*TCPConn)
}

func Addr(ip string, port uint16) string { return fmt.Sprintf("%s:%d", ip, port) }

func (self *TCPClient) ConnectToSvr(addr string, meta *meta.Meta) {
	self.addr = addr
	go self.connectRoutine(meta) //会断线后自动重连
}
func (self *TCPClient) connectRoutine(meta *meta.Meta) {
	regMsg := common.NewNetPackCap(32)
	regMsg.SetOpCode(enum.Rpc_regist)
	meta.DataToBuf(regMsg)
	for {
		if self.connect() {
			//Notice: 这里不能用CallRpc，非线程安全的
			firstMsg := make([]byte, 2+4) //tcp层的包，头两个字节是长度
			binary.LittleEndian.PutUint32(firstMsg[2:], self.connId)
			self.TcpConn.WriteBuf(firstMsg)
			self.TcpConn.WriteMsg(regMsg)
			if self.OnConnect != nil {
				self.OnConnect(self.TcpConn)
			}
			go self.TcpConn.readRoutine()
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
		self.TcpConn = newTCPConn(conn)
		self.TcpConn.UserPtr = self
	} else {
		//断线重连的新连接标记得重置，否则tcpConn.readRoutine.readLoop会直接break
		self.TcpConn.resetConn(conn)
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
