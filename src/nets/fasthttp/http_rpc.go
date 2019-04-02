package fasthttp

import (
	"common"
	"common/std/compress"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	http "github.com/valyala/fasthttp"
	"svr_client/test/qps"
)

var (
	G_HandleFunc [enum.RpcEnumCnt]func(req, ack *common.NetPack)
	G_Intercept  func(req, ack *common.NetPack, clientIp string) bool
)

// ------------------------------------------------------------
//! system rpc
func CallRpc(addr string, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	req := common.NewNetPackCap(64)
	req.SetOpCode(rid)
	sendFun(req)
	if buf := PostReq(addr+"/client_rpc", req.Data()); buf != nil && recvFun != nil {
		if ack := common.NewNetPack(compress.Decompress(buf)); ack != nil {
			recvFun(ack)
		}
	}
	req.Free()
}
func _HandleRpc(ctx *http.RequestCtx) {
	req := common.NewNetPack(ctx.Request.Body())
	ack := common.NewNetPackCap(128)
	defer func() {
		compress.CompressTo(ack.Data(), ctx)
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
	//defer func() {//库已经有recover了，见net/http/server.go:1918
	//	if r := recover(); r != nil {
	//		gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
	//	}
	//	ack.Free()
	//}()
	if G_Intercept != nil { //拦截器
		ip := ctx.RemoteIP().String()
		if G_Intercept(req, ack, ip) {
			gamelog.Info("Intercept msg:%d ip:%s", msgId, ip)
			return
		}
	}
	if handler := G_HandleFunc[msgId]; handler != nil {
		handler(req, ack)
	} else {
		gamelog.Error("Msg(%d) Not Regist", msgId)
	}
}

// ------------------------------------------------------------
//! player rpc
type PlayerRpc struct {
	url       string
	AccountId uint32
}

func RegHandlePlayerRpc(cb func(*http.RequestCtx)) {
	HandleFunc("/player_rpc", cb)
}

func NewPlayerRpc(addr string, accountId uint32) *PlayerRpc {
	return &PlayerRpc{addr + "/player_rpc", accountId}
}
func (self *PlayerRpc) CallRpc(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	req := common.NewNetPackCap(64)
	req.SetOpCode(rid)
	req.SetReqIdx(self.AccountId)
	sendFun(req)
	if buf := PostReq(self.url, req.Data()); buf != nil {
		if ack := common.NewNetPack(compress.Decompress(buf)); ack != nil {
			if recvFun != nil {
				recvFun(ack)
			}
			_RecvHttpSvrData(ack) //服务器主动下发的数据
		}
	}
	req.Free()
}
func _RecvHttpSvrData(buf *common.NetPack) {
	//对应于 http_to_client.go
}
