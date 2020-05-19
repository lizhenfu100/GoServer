package http

import (
	"common"
	"common/std/compress"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"io"
	"nets/rpc"
	"svr_client/test/qps"
	"sync/atomic"
)

func CallRpc(addr string, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	req := common.NewNetPackCap(32)
	req.SetMsgId(rid)
	sendFun(req)
	buf := Client.PostReq(addr+"/client_rpc", req.Data())
	if buf != nil && recvFun != nil {
		if ack := common.ToNetPack(compress.Decompress(buf)); ack != nil {
			recvFun(ack)
		}
	}
	req.Free()
}
func HandleRpc(request []byte, w io.Writer, clientIp string) { //G_Intercept==nil,外界不必取ip,传空即可
	if conf.TestFlag_CalcQPS {
		qps.AddQps()
	}
	req := common.ToNetPack(request)
	if req == nil {
		gamelog.Error("invalid req: %v", request)
		return
	}
	if msgId := req.GetMsgId(); msgId < enum.RpcEnumCnt {
		//gamelog.Debug("HttpMsg:%d, len:%d", msgId, req.BodySize())
		ack := common.NewNetPackCap(64)
		if p := Intercept(); p == nil || !p(req, ack, clientIp) { //拦截器
			if msgFunc := rpc.G_HandleFunc[msgId]; msgFunc != nil {
				msgFunc(req, ack, nil)
			} else {
				gamelog.Error("Msg(%d) Not Regist", msgId)
			}
		}
		compress.CompressTo(ack.Data(), w)
		ack.Free()
	}
	req.Free()
}

// ------------------------------------------------------------
// -- 拦截器
type intercept func(req, ack *common.NetPack, clientIp string) bool

var _intercept atomic.Value

func SetIntercept(f intercept) {
	if Intercept() != nil {
		panic("intercept repeat")
	}
	_intercept.Store(f)
}
func Intercept() intercept { v, _ := _intercept.Load().(intercept); return v }
