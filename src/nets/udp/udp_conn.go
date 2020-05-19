package udp

import (
	"common"
	"common/safe"
	"gamelog"
	"net"
	"nets/rpc"
	"sync/atomic"
	"time"
)

const (
	Writer_Cap = 8 * 1024
)

type UDPConn struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	writer   safe.ChanByte
	_isClose bool
	user     atomic.Value
}

var _connmap = map[string]*UDPConn{}

func Session(p *net.UDPAddr, c *net.UDPConn) *UDPConn {
	conn := _connmap[p.String()]
	if conn == nil {
		conn = newConn(c, p)
		_connmap[p.String()] = conn
		go conn.writeLoop()
	}
	return conn
}

func (self *UDPConn) readLoop() {
	for req, q := common.NewByteBufferLen(2048), &rpc.G_RpcQueue; ; {
		if n, addr, e := self.conn.ReadFromUDP(req.Data()); e != nil {
			gamelog.Debug(e.Error())
			break
		} else if n > 0 {
			q.Insert(Session(addr, self.conn), req)
		}
	}
}
func (self *UDPConn) Connect(addr string) {
	self.addr, _ = net.ResolveUDPAddr("udp", addr)
	conn, err := net.DialUDP("udp", nil, self.addr)
	if err != nil {
		gamelog.Error("Connect %s: %s", addr, err.Error())
		return
	}
	Session(self.addr, conn)
	self.conn = conn
	go self.readLoop()
}

// ------------------------------------------------------------
func newConn(conn *net.UDPConn, addr *net.UDPAddr) *UDPConn {
	self := new(UDPConn)
	self.writer.Init(2048)
	self.reset(conn, addr)
	return self
}
func (self *UDPConn) reset(conn *net.UDPConn, addr *net.UDPAddr) {
	self.conn = conn
	self.addr = addr
	self.writer.IsStop = false
	self._isClose = false
}
func (self *UDPConn) Close() {
	if self._isClose {
		return
	}
	self.conn.Close()
	self._isClose = true
}
func (self *UDPConn) IsClose() bool { return self._isClose }

func (self *UDPConn) GetUser() interface{}  { return self.user.Load() }
func (self *UDPConn) SetUser(v interface{}) { self.user.Store(v) }

func (self *UDPConn) CallRpc(msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	req := common.NewNetPackCap(32)
	rpc.MakeReq(req, msgId, sendFun, recvFun)
	self.WriteMsg(req)
	req.Free()
}
func (self *UDPConn) WriteMsg(msg *common.NetPack) {
	if self.writer.AddMsg(msg.Data(), msg.Size()) > Writer_Cap {
		self.Close()
	}
}
func (self *UDPConn) WriteBuf(buf []byte) { self.writer.Add(buf) }
func (self *UDPConn) writeLoop() { //TODO:zhoumf:UDP并包，一次发个MTU大小
loop:
	for {
		b := self.writer.WaitGet() //block
		for pos, total := 0, len(b); ; {
			if n, e := self.conn.WriteToUDP(b[pos:], self.addr); e != nil {
				gamelog.Debug(e.Error())
				break loop
			} else if pos += n; pos == total {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	self.Close()
}
