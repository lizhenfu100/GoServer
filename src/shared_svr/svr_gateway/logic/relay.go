package logic

import (
	"common"
	"common/assert"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/http"
)

func RelayHttp(req, ack *common.NetPack, _ common.Conn) {
	Rpc_gateway_relay(req, func(backBuf *common.NetPack) {
		ack.ResetHead(backBuf)
		ack.WriteBuf(backBuf.LeftBuf())
	})
}
func RelayTcp(req, _ *common.NetPack, conn common.Conn) {
	oldReqKey := req.GetReqKey()
	Rpc_gateway_relay(req, func(backBuf *common.NetPack) {
		backBuf.SetReqKey(oldReqKey)
		conn.WriteMsg(backBuf) //异步回调，不能直接用ack
	})
}
func Rpc_gateway_relay(req *common.NetPack, recvFun func(*common.NetPack)) {
	rpcId := req.ReadUInt16()     //目标rpc
	accountId := req.ReadUInt32() //目标玩家
	args := req.LeftBuf()
	assert.True(accountId > 0 && netConfig.HashGatewayID(accountId) == meta.G_Local.SvrID)
	errcode := byte(common.Err_offline)
	switch module := enum.GetRpcModule(rpcId); module {
	case "client":
		if p := GetClientConn(accountId); p != nil {
			p.CallRpc(rpcId, func(buf *common.NetPack) {
				buf.WriteBuf(args)
			}, recvFun)
			errcode = 0
		}
	case "game":
		if id := GetGameId(accountId); id > 0 {
			if p, ok := netConfig.GetGameRpc(id); ok {
				p.CallRpc(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
					buf.WriteUInt16(rpcId)
					buf.WriteUInt32(accountId)
					buf.WriteBuf(args)
				}, recvFun)
				errcode = 0
			}
		}
	case "save": //附属svr_game
		if id := GetGameId(accountId); id > 0 {
			if p := meta.GetMeta("game", id); p != nil {
				http.CallRpc(http.Addr(p.IP, meta.KSavePort), rpcId, func(buf *common.NetPack) {
					buf.WriteBuf(args)
				}, recvFun)
				errcode = 0
			}
		}
	default:
		var pMeta *meta.Meta
		if accountId > 0 {
			pMeta = meta.GetByMod(module, accountId)
		} else {
			pMeta = meta.GetByRand(module)
		}
		if pMeta != nil {
			if p, ok := netConfig.GetRpc1(pMeta); ok {
				p.CallRpc(rpcId, func(buf *common.NetPack) {
					buf.WriteBuf(args)
				}, recvFun)
				errcode = 0
			}
		}
	}
	if errcode != 0 {
		ReportErr(recvFun, errcode)
	}
}
func ReportErr(f func(*common.NetPack), e byte) {
	ack := common.NewNetPackCap(16)
	ack.SetType(e)
	f(ack)
	ack.Free()
}

// ------------------------------------------------------------
// 转发http
//var _sdk http.Handler
//func relaySdk(w http.ResponseWriter, r *http.Request) { _sdk.ServeHTTP(w, r) }
//func init() {
//	_sdk = &httputil.ReverseProxy{Director: func(r *http.Request) {
//		if p := meta.GetByRand("sdk"); p != nil {
//			r.URL.Scheme = "http"
//			r.URL.Host = fmt.Sprintf("%s:%d", p.IP, p.HttpPort)
//		}
//	}}
//	http.HandleFunc("/pre_buy_request", relaySdk)
//	http.HandleFunc("/query_order", relaySdk)
//	http.HandleFunc("/confirm_order", relaySdk)
//	http.HandleFunc("/query_order_unfinished", relaySdk)
//}
