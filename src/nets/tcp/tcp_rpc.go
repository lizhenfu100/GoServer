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
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"runtime/debug"
	"svr_client/test/qps"
	"sync"
	"sync/atomic"
)

type objMsg struct {
	conn *TCPConn
	msg  *common.NetPack
}
type RpcQueue struct {
	queue    chan objMsg //safe.Pipe
	sendBuf  *common.NetPack
	backBuf  *common.NetPack
	reqIdx   uint32
	response sync.Map //<reqkey uint64, rpcRecv func(*common.NetPack) >

	//处理与玩家强绑定的rpc
	_playerRpc func(req, ack *common.NetPack, conn *TCPConn) bool
}

func (self *RpcQueue) Init() {
	//self.queue.Init(cap)
	self.queue = make(chan objMsg, Msg_Queue_Cap)
	self.sendBuf = common.NewNetPackCap(64)
	self.backBuf = common.NewNetPackCap(64)
}

//【单线程：req每次new，ack同一个】【多线程：thread local】
//Notice：本函数过后，req、ack生命周期结束。check_closure.go 检查req被闭包
func (self *RpcQueue) _Handle(conn *TCPConn, req, ack *common.NetPack) {
	if conf.TestFlag_CalcQPS {
		qps.AddQps()
	}
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("%v: %s", r, debug.Stack())
		}
	}()
	if req.GetType() == 0 {
		if msgId := req.GetMsgId(); msgId < enum.RpcEnumCnt {
			ack.ResetHead(req)
			if self._playerRpc != nil && self._playerRpc(req, ack, conn) { //1、先处理玩家rpc(与player指针强关联)
				//gamelog.Debug("PlayerMsg:%d, len:%d", msgId, req.BodySize())
			} else if msgFunc := G_HandleFunc[msgId]; msgFunc != nil { //2、普通类型rpc(与连接关联的)
				//gamelog.Debug("TcpMsg:%d, len:%d", msgId, req.BodySize())
				msgFunc(req, ack, conn)
			} else {
				gamelog.Error("Msg(%d) Not Regist", msgId)
			}
			if ack.GetType() > 0 {
				conn.WriteMsg(ack)
			} else if ack.BodySize() > 0 {
				ack.SetType(common.Type_ack)
				conn.WriteMsg(ack)
			}
		}
	} else { //3、rpc回复(自己发起的调用，对方回复的数据)
		reqKey := req.GetReqKey()
		if rpcRecv, ok := self.response.Load(reqKey); ok {
			rpcRecv.(func(*common.NetPack))(req)
			self.response.Delete(reqKey)
		}
	}
}
func RegHandlePlayerRpc(cb func(req, ack *common.NetPack, conn *TCPConn) bool) {
	G_RpcQueue._playerRpc = cb // 与玩家强绑定的rpc
}

//Notice：非线程安全的，仅供主逻辑线程调用，内部操作的同个sendBuf，多线程下须每次new新的
func (self *TCPConn) CallRpc(msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	q := &G_RpcQueue
	assert.True(q.sendBuf.GetMsgId() == 0)
	q.sendBuf.SetMsgId(msgId) //中途不能再CallRpc，同个sendBuf被覆盖
	q.sendBuf.SetReqIdx(atomic.AddUint32(&q.reqIdx, 1))
	if recvFun != nil {
		q.response.Store(q.sendBuf.GetReqKey(), recvFun)
	}
	sendFun(q.sendBuf)
	self.WriteMsg(q.sendBuf)
	q.sendBuf.ClearBody()
	q.sendBuf.SetMsgId(0)
}
func (self *TCPConn) CallRpcSafe(msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	q := &G_RpcQueue
	req := common.NewNetPackCap(32)
	req.SetMsgId(msgId)
	req.SetReqIdx(atomic.AddUint32(&q.reqIdx, 1))
	if recvFun != nil {
		q.response.Store(req.GetReqKey(), recvFun)
	}
	sendFun(req)
	self.WriteMsg(req)
	req.Free()
}
