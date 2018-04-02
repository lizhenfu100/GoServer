/***********************************************************************
* @ 单线程Rpc队列
* @ brief
	1、转接网络层buffer中的数据，缓存，以待主逻辑循环处理【因需等待主逻辑，消息响应会加大延时】

    2、同名rpc混乱：client rpc server且有回包；若server那边也有个同名rpc client，那client就不好区分底层收到的包，是自己rpc的回复，还是对方主动rpc

	3、远程调用其它模块的rpc，应是本模块未声明实现的，避免同名rpc的混乱

* @ author zhoumf
* @ date 2017-12-27
************************************************************************/
package tcp

import (
	"common"
	"gamelog"
	"runtime/debug"
	"sync/atomic"
)

type objMsg struct {
	conn *TCPConn
	msg  *common.NetPack
}
type RpcQueue struct {
	//queue      *safe.SafeQueue //{objMsg}
	queue      chan objMsg
	sendBuffer *common.NetPack
	backBuffer *common.NetPack
	reqIdx     uint32
	response   map[uint64]func(*common.NetPack)

	//处理与玩家强绑定的rpc
	_handlePlayerRpc func(req, ack *common.NetPack, conn *TCPConn) bool
}

func NewRpcQueue(cap uint32) *RpcQueue {
	self := new(RpcQueue)
	//self.queue = safe.NewQueue(cap)
	self.queue = make(chan objMsg, Msg_Queue_Cap)
	self.sendBuffer = common.NewNetPackCap(128)
	self.backBuffer = common.NewNetPackCap(128)
	self.response = make(map[uint64]func(*common.NetPack))
	return self
}

func (self *RpcQueue) Insert(conn *TCPConn, msg *common.NetPack) {
	//self.queue.Put(objMsg{conn, msg})
	select {
	case self.queue <- objMsg{conn, msg}:
	default:
		gamelog.Error("RpcQueue Insert: channel full")
		conn.Close()
	}
}

//主循环，每帧调一次
func (self *RpcQueue) Update() {
	//if v, ok, _ := self.queue.Get(); ok {
	//	self._Handle(v.(objMsg).conn, v.(objMsg).msg)
	//	v.(objMsg).msg.Free()
	//}
	for {
		select {
		case v := <-self.queue:
			self._Handle(v.conn, v.msg)
			v.msg.Free()
		default:
			return
		}
	}
}
func (self *RpcQueue) Loop() { //死循环，阻塞等待
	for {
		select {
		case v := <-self.queue:
			self._Handle(v.conn, v.msg)
			v.msg.Free()
		default:
			v := <-self.queue
			self._Handle(v.conn, v.msg)
			v.msg.Free()
		}
	}
}

//Notice：非线程安全的，将 response 改为 sync.Map 才安全的
func (self *RpcQueue) _Handle(conn *TCPConn, msg *common.NetPack) {
	msgId := msg.GetOpCode()
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
		}
	}()
	gamelog.Debug("Recv TcpMsgId:%d, len:%d", msgId, msg.Size())

	self.backBuffer.ResetHead(msg)

	if self._handlePlayerRpc != nil && self._handlePlayerRpc(msg, self.backBuffer, conn) {
		if self.backBuffer.BodySize() > 0 {
			conn.WriteMsg(self.backBuffer)
		}
	} else if msgFunc := G_HandleFunc[msgId]; msgFunc != nil {
		msgFunc(msg, self.backBuffer, conn)
		if self.backBuffer.BodySize() > 0 {
			conn.WriteMsg(self.backBuffer)
		}
	} else if rpcRecv, ok := self.response[msg.GetReqKey()]; ok {
		rpcRecv(msg)
		delete(self.response, msg.GetReqKey())
	} else {
		gamelog.Error("msgid[%d] havn't msg handler!!", msgId)
	}
}

// 处理与玩家强绑定的rpc
func RegHandlePlayerRpc(cb func(req, ack *common.NetPack, conn *TCPConn) bool) {
	G_RpcQueue._handlePlayerRpc = cb
}

//Notice：非线程安全的，仅供主逻辑线程调用，内部操作的同个sendBuffer，多线程下须每次new新的
func (self *RpcQueue) CallRpc(conn *TCPConn, msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	common.Assert(G_HandleFunc[msgId] == nil && self.sendBuffer.GetOpCode() == 0)
	//CallRpc中途不能再CallRpc
	//gamelog.Error("[%d] Server and Client have the same Rpc or Repeat CallRpc", msgID)

	self.sendBuffer.SetOpCode(msgId)
	self.sendBuffer.SetReqIdx(atomic.AddUint32(&self.reqIdx, 1))
	if recvFun != nil {
		self.response[self.sendBuffer.GetReqKey()] = recvFun
	}
	sendFun(self.sendBuffer)
	conn.WriteMsg(self.sendBuffer)
	self.sendBuffer.Clear()
}
