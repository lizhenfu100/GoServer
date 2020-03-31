package logic

import (
	"common"
	"common/assert"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/http"
)

func Rpc_gateway_relay(req *common.NetPack, recvFun func(*common.NetPack)) {
	rpcId := req.ReadUInt16()     //目标rpc
	accountId := req.ReadUInt32() //目标玩家
	args := req.LeftBuf()
	assert.True(accountId > 0 && netConfig.HashGatewayID(accountId) == meta.G_Local.SvrID)
	errcode := byte(common.Err_offline)
	switch module := enum.GetRpcModule(rpcId); module {
	case "client":
		if p := GetClientConn(accountId); p != nil {
			p.CallRpcSafe(rpcId, func(buf *common.NetPack) {
				buf.WriteBuf(args)
			}, recvFun)
			errcode = 0
		}
	case "game":
		if id := GetGameId(accountId); id > 0 {
			if p, ok := netConfig.GetGameRpc(id); ok {
				p.CallRpcSafe(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
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
				p.CallRpcSafe(rpcId, func(buf *common.NetPack) {
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
