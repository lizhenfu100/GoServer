package tcp

import (
	"common"
	"encoding/binary"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"net"
	"netConfig/meta"
	"sync/atomic"
	"time"
)

type TCPClient struct {
	addr      string
	connId    uint32
	Conn      *TCPConn
	onConnect func(*TCPConn)
	_isClose  int32
}

func Addr(ip string, port uint16) string { return fmt.Sprintf("%s:%d", ip, port) }

func (self *TCPClient) Connect(addr string, cb func(*TCPConn)) {
	if self.addr != addr {
		self.addr = addr
		self.onConnect = cb
		if self.Conn == nil {
			go self.connectRoutine() //会断线后自动重连
		} else {
			self.Conn.Close() //关老连接，重连新地址
		}
	}
}
func (self *TCPClient) connectRoutine() {
	firstMsg := make([]byte, 4) //connId
	if meta.G_Local != nil {    //后台节点间的连接，发起注册
		regMsg := common.NewNetPackCap(128)
		regMsg.SetMsgId(enum.Rpc_regist)
		meta.G_Local.DataToBuf(regMsg)
		//将regMsg追加到firstMsg之后，须满足tcp包格式
		firstMsg = append(firstMsg, byte(regMsg.Size()), byte(regMsg.Size()>>8))
		firstMsg = append(firstMsg, regMsg.Data()...)
		regMsg.Free()
	}
	for atomic.LoadInt32(&self._isClose) == 0 {
		if self.connect() {
			binary.LittleEndian.PutUint32(firstMsg, self.connId)
			self.Conn.WriteBuf(firstMsg) //Notice: 不能用CallRpc，非线程安全的
			//if self.onConnect != nil { //Notice：放这里，回调就是多线程执行的了，健壮性低
			//	self.onConnect(self.Conn)
			//}
			go self.Conn.writeLoop()
			self.Conn.readLoop() //goroutine会阻塞在这里
		}
		time.Sleep(3 * time.Second)
	}
}
func (self *TCPClient) connect() bool {
	conn, err := net.Dial("tcp", self.addr)
	if err != nil || conn == nil {
		gamelog.Error("connect to %s :%s", self.addr, err.Error())
		return false
	}
	if self.Conn == nil {
		self.Conn = newTCPConn(conn)
		self.Conn.SetUser(self)
	} else {
		//断线重连的新连接标记得重置，否则tcpConn.readLoop会直接break
		self.Conn.resetConn(conn)
	}
	return true
}
func (self *TCPClient) Close() {
	atomic.StoreInt32(&self._isClose, 1)
	if self.Conn != nil {
		self.Conn.Close()
		self.Conn = nil
	}
}
func _Rpc_svr_accept(req, ack *common.NetPack, conn common.Conn) {
	self := conn.GetUser().(*TCPClient)
	self.connId = req.ReadUInt32()
	if self.onConnect != nil {
		self.onConnect(self.Conn)
	}
}
