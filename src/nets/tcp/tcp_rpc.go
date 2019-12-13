/***********************************************************************
* @ 单线程Rpc队列
* @ brief
	1、转接网络层buffer中的数据，缓存，以待主逻辑循环处理【因需等待主逻辑，消息响应会加大延时】

    2、同名rpc混乱：
		· client call server且有回包
		· client 本地也有个同名 rpc
		· client 不好区分收包，是自己rpc的回复，还是对方call自己的

	3、远程调用其它模块的rpc，应是本模块未声明实现的，避免同名rpc的混乱

* @ author zhoumf
* @ date 2017-12-27
************************************************************************/
package tcp

import (
	"common"
	"common/assert"
	"gamelog"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type objMsg struct {
	conn *TCPConn
	msg  *common.NetPack
}
type RpcQueue struct {
	//queue      safe.MultiQueue //{objMsg}
	queue      chan objMsg
	sendBuffer *common.NetPack
	backBuffer *common.NetPack
	_reqIdx    uint32
	response   sync.Map //<reqkey uint64, rpcRecv func(*common.NetPack) >

	//处理与玩家强绑定的rpc
	_handlePlayerRpc func(req, ack *common.NetPack, conn *TCPConn) bool
}

func (self *RpcQueue) Init(cap uint32) {
	//self.queue.Init(cap)
	self.queue = make(chan objMsg, cap)
	self.sendBuffer = common.NewNetPackCap(128)
	self.backBuffer = common.NewNetPackCap(128)
}

func (self *RpcQueue) Update() { //主循环，每帧调一次
	//for _, v := range self.queue.Get() {
	//	self._Handle(v.(objMsg).conn, v.(objMsg).msg)
	//	v.(objMsg).msg.Free()
	//}
	for {
		select {
		case v := <-self.queue:
			self._Handle(v.conn, v.msg, self.backBuffer)
		default:
			return
		}
	}
}
func (self *RpcQueue) Loop() { //死循环，阻塞等待
	for {
		v := <-self.queue
		self._Handle(v.conn, v.msg, self.backBuffer)
	}
}

// 单线程用self.backBuffe，多线程每次new新的
func (self *RpcQueue) _Handle(conn *TCPConn, msg, backBuf *common.NetPack) {
	msgId := msg.GetMsgId()
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
		}
	}()
	backBuf.ResetHead(msg)
	//1、先处理玩家rpc(与player指针强关联)
	if self._handlePlayerRpc != nil && self._handlePlayerRpc(msg, backBuf, conn) {
		gamelog.Debug("PlayerMsg:%d, len:%d", msgId, msg.BodySize())
		if backBuf.BodySize() > 0 || backBuf.GetType() > 0 {
			conn.WriteMsg(backBuf)
		}
		//2、普通类型rpc(与连接关联的)
	} else if msgFunc := G_HandleFunc[msgId]; msgFunc != nil {
		gamelog.Debug("TcpMsg:%d, len:%d", msgId, msg.BodySize())
		msgFunc(msg, backBuf, conn)
		if backBuf.BodySize() > 0 || backBuf.GetType() > 0 {
			conn.WriteMsg(backBuf)
		}
		//3、rpc回复(自己发起的调用，对方回复的数据)
	} else if rpcRecv, ok := self.response.Load(msg.GetReqKey()); ok {
		gamelog.Debug("ResponseMsg:%d, len:%d", msgId, msg.BodySize())
		rpcRecv.(func(*common.NetPack))(msg)
		self.response.Delete(msg.GetReqKey())
	} else {
		gamelog.Error("Msg(%d) Not Regist", msgId)
	}
	msg.Free()
}
func RegHandlePlayerRpc(cb func(req, ack *common.NetPack, conn *TCPConn) bool) {
	G_RpcQueue._handlePlayerRpc = cb // 与玩家强绑定的rpc
}

//Notice：非线程安全的，仅供主逻辑线程调用，内部操作的同个sendBuffer，多线程下须每次new新的
func (self *TCPConn) CallRpc(msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	g := &G_RpcQueue
	assert.True(G_HandleFunc[msgId] == nil && g.sendBuffer.GetMsgId() == 0)
	//CallRpc中途不能再CallRpc
	//gamelog.Error("[%d] Server and Client have the same Rpc or Repeat CallRpc", msgID)
	g.sendBuffer.SetMsgId(msgId)
	g.sendBuffer.SetReqIdx(atomic.AddUint32(&g._reqIdx, 1))
	if recvFun != nil {
		g.response.Store(g.sendBuffer.GetReqKey(), recvFun)
	}
	sendFun(g.sendBuffer)
	self.WriteMsg(g.sendBuffer)
	g.sendBuffer.Clear()
}
func (self *TCPConn) CallRpcSafe(msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	g := &G_RpcQueue
	assert.True(G_HandleFunc[msgId] == nil)
	req := common.NewNetPackCap(64)
	req.SetMsgId(msgId)
	req.SetReqIdx(atomic.AddUint32(&g._reqIdx, 1))
	if recvFun != nil {
		g.response.Store(req.GetReqKey(), recvFun)
	}
	sendFun(req)
	self.WriteMsg(req)
	req.Free()
}
