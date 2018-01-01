/***********************************************************************
* @ 单线程Rpc队列
* @ brief
	1、转接网络层buffer中的数据，缓存，以待主逻辑循环处理

    2、“同名rpc混乱”：client rpc server且有回包；若server那边也有个同名rpc client，那client就不好区分底层收到的包，是自己rpc的回复，还是对方主动rpc

	3、远程调用其它模块的rpc，应是本模块未声明实现的，避免同名rpc的混乱

* @ author zhoumf
* @ date 2017-12-27
************************************************************************/
package tcp

import (
	"common"
	"common/safe"
	"gamelog"
	"runtime/debug"
	"sync/atomic"
)

type objMsg struct {
	conn *TCPConn
	msg  *common.NetPack
}
type RpcQueue struct {
	queue      *safe.SafeQueue //{objMsg}
	sendBuffer *common.NetPack
	backBuffer *common.NetPack
	reqIdx     uint32
	response   map[uint64]func(*common.NetPack)
}

func NewRpcQueue(cap uint32) *RpcQueue {
	self := new(RpcQueue)
	self.queue = safe.NewQueue(cap)
	self.sendBuffer = common.NewNetPackCap(128)
	self.backBuffer = common.NewNetPackCap(128)
	self.response = make(map[uint64]func(*common.NetPack))
	return self
}

func (self *RpcQueue) Insert(conn *TCPConn, msg *common.NetPack) {
	self.queue.Put(objMsg{conn, msg})
}

//主循环，每帧调一次
func (self *RpcQueue) Update() {
	if v, ok, _ := self.queue.Get(); ok {
		self._Handle(v.(objMsg).conn, v.(objMsg).msg)
		v.(objMsg).msg.Free()
	}
}

//Notice：该函数非线程安全的，将 response 改为 sync.Map 才安全的
func (self *RpcQueue) _Handle(conn *TCPConn, msg *common.NetPack) {
	msgId := msg.GetOpCode()
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
		}
	}()
	if msgFunc := G_HandleFunc[msgId]; msgFunc != nil {
		//gamelog.Debug("TcpMsgId:%d, len:%d", msgId, msg.Size())
		self.backBuffer.ResetHead(msg)
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

//Notice：仅供主逻辑线程调用，内部重复操作的同个sendBuffer，多线程下须每次new新的
func (self *RpcQueue) CallRpc(conn *TCPConn, msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	common.Assert(G_HandleFunc[msgId] == nil && self.sendBuffer.GetOpCode() == 0)
	//gamelog.Error("[%d] Server and Client have the same Rpc or Repeat CallRpc", msgID)

	self.sendBuffer.SetOpCode(msgId)
	self.sendBuffer.SetReqIdx(atomic.AddUint32(&self.reqIdx, 1))
	sendFun(self.sendBuffer)
	conn.WriteMsg(self.sendBuffer)
	self.sendBuffer.Clear()

	if recvFun != nil {
		self.response[self.sendBuffer.GetReqKey()] = recvFun
	}
}
