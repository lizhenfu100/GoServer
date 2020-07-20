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
package rpc

import (
	. "common"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"runtime/debug"
	"svr_client/test/qps"
	"sync"
	"sync/atomic"
)

var (
	G_HandleFunc [enum.RpcEnumCnt]func(req, ack *NetPack, conn Conn)
	G_RpcQueue   RpcQueue
)

type RpcQueue struct {
	reqIdx   uint32
	response sync.Map //<reqkey uint64, rpcRecv func(*common.NetPack) >
	//处理与玩家强绑定的rpc
	_playerRpc func(req, ack *NetPack, conn Conn) bool
}
type objMsg struct {
	conn Conn
	msg  *NetPack
}

func MakeReq(req *NetPack, msgId uint16, sendFun, recvFun func(*NetPack)) {
	q := &G_RpcQueue
	req.SetMsgId(msgId)
	req.SetReqIdx(atomic.AddUint32(&q.reqIdx, 1))
	if recvFun != nil {
		q.response.Store(req.GetReqKey(), recvFun)
	}
	sendFun(req)
}

//【单线程：req每次new，ack同一个】【多线程：thread local】
//Notice：本函数过后，req、ack生命周期结束。check_closure.go 检查req被闭包
func (self *RpcQueue) Handle(conn Conn, req, ack *NetPack) {
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
			//gamelog.Debug("Msg:%d, len:%d", msgId, req.BodySize())
			ack.ResetHead(req)
			if self._playerRpc != nil && self._playerRpc(req, ack, conn) { //1、先处理玩家rpc(与player指针强关联)

			} else if msgFunc := G_HandleFunc[msgId]; msgFunc != nil { //2、普通类型rpc(与连接关联的)
				msgFunc(req, ack, conn)
			} else {
				gamelog.Error("Msg(%d) Not Regist", msgId)
			}
			if ack.GetType() > 0 {
				conn.WriteMsg(ack)
			} else if ack.BodySize() > 0 {
				ack.SetType(Type_ack)
				conn.WriteMsg(ack)
			}
		}
	} else { //3、rpc回复(自己发起的调用，对方回复的数据)
		reqKey := req.GetReqKey()
		if rpcRecv, ok := self.response.Load(reqKey); ok {
			rpcRecv.(func(*NetPack))(req)
			self.response.Delete(reqKey)
		} else {
			gamelog.Error("Msg(%d) None response: %d", req.GetMsgId(), req.GetType())
		}
	}
}
func RegHandlePlayerRpc(cb func(req, ack *NetPack, conn Conn) bool) {
	G_RpcQueue._playerRpc = cb // 与玩家强绑定的rpc
}
