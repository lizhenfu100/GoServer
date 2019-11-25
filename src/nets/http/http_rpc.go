package http

import (
	"common"
	"common/std/compress"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"io"
	"svr_client/test/qps"
	"sync/atomic"
)

var G_HandleFunc [enum.RpcEnumCnt]func(req, ack *common.NetPack)

func CallRpc(addr string, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	req := common.NewNetPackCap(64)
	req.SetOpCode(rid)
	sendFun(req)
	buf := Client.PostReq(addr+"/client_rpc", req.Data())
	if buf != nil && recvFun != nil {
		if ack := common.NewNetPack(compress.Decompress(buf)); ack != nil {
			recvFun(ack)
		}
	}
	req.Free()
}
func HandleRpc(request []byte, w io.Writer, clientIp string) { //G_Intercept==nil,外界不必取ip,传空即可
	req := common.NewNetPack(request)
	if req == nil {
		gamelog.Error("invalid req: %v", request)
		return
	}
	ack := common.NewNetPackCap(128)
	defer func() {
		compress.CompressTo(ack.Data(), w)
		ack.Free()
		req.Free()
	}()
	msgId := req.GetOpCode()
	if msgId >= enum.RpcEnumCnt {
		gamelog.Error("Msg(%d) Not Regist", msgId)
		return
	}
	gamelog.Debug("HttpMsg:%d, len:%d", msgId, req.Size())

	if conf.TestFlag_CalcQPS {
		qps.AddQps()
	}
	if p := Intercept(); p != nil && p(req, ack, clientIp) { //拦截器
		return
	}
	if handler := G_HandleFunc[msgId]; handler != nil {
		handler(req, ack)
	} else {
		gamelog.Error("Msg(%d) Not Regist", msgId)
	}
}

// ------------------------------------------------------------
// -- 拦截器
type intercept func(req, ack *common.NetPack, clientIp string) bool

var g_intercept atomic.Value

func SetIntercept(f intercept) { g_intercept.Store(f) }
func Intercept() intercept     { v, _ := g_intercept.Load().(intercept); return v }
