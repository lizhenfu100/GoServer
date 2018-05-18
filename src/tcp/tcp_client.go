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
	OnConnect func(*TCPConn)
	_isLoop   int32
}

func Addr(ip string, port uint16) string { return fmt.Sprintf("%s:%d", ip, port) }

func (self *TCPClient) ConnectToSvr(addr string, myMeta *meta.Meta) {
	if atomic.LoadInt32(&self._isLoop) == 0 {
		self.addr = addr
		go self.connectRoutine(myMeta) //会断线后自动重连
	}
}
func (self *TCPClient) connectRoutine(myMeta *meta.Meta) {
	regMsg := common.NewNetPackCap(32)
	regMsg.SetOpCode(enum.Rpc_regist)
	myMeta.DataToBuf(regMsg)

	atomic.StoreInt32(&self._isLoop, 1)
	for {
		if self.connect() {
			//Notice: 这里不能用CallRpc，非线程安全的
			firstMsg := make([]byte, 2+4) //tcp层的包，头两个字节是长度
			binary.LittleEndian.PutUint32(firstMsg[2:], self.connId)
			self.Conn.WriteBuf(firstMsg)
			self.Conn.WriteMsg(regMsg)
			//if self.OnConnect != nil { //Notice：放这里，回调就是多线程执行的了，健壮性低
			//	self.OnConnect(self.Conn)
			//}
			go self.Conn.readRoutine()
			self.Conn.writeRoutine() //goroutine会阻塞在这里
		}

		//有zookeeper管理节点连接，无需自动重连
		if meta.GetMeta("zookeeper", 0) != nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	atomic.StoreInt32(&self._isLoop, 0)
}
func (self *TCPClient) connect() bool {
	conn, err := net.Dial("tcp", self.addr)
	if err != nil || conn == nil {
		gamelog.Error("connect to %s :%s", self.addr, err.Error())
		return false
	}
	if self.Conn == nil {
		self.Conn = newTCPConn(conn)
		self.Conn.UserPtr = self
	} else {
		//断线重连的新连接标记得重置，否则tcpConn.readRoutine.readLoop会直接break
		self.Conn.resetConn(conn)
	}
	return true
}
func (self *TCPClient) Close() {
	self.Conn.Close()
	self.Conn = nil
}
func _Rpc_svr_accept(req, ack *common.NetPack, conn *TCPConn) {
	self := conn.UserPtr.(*TCPClient)
	self.connId = req.ReadUInt32()
	if self.OnConnect != nil {
		self.OnConnect(self.Conn)
	}
}
