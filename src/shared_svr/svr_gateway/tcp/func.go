package tcp

import (
	"common"
	"generate_out/err"
	"generate_out/rpc/enum"
	"netConfig/meta"
	"nets/tcp"
	"shared_svr/svr_gateway/logic"
)

func Rpc_gateway_login(req, ack *common.NetPack, client *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()

	if logic.CheckToken(accountId, token) {
		client.UserPtr = accountId
		logic.AddClientConn(accountId, client)
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Token_verify_err)
	}
}
func Rpc_gateway_login_token(req, ack *common.NetPack, conn *tcp.TCPConn) {
	logic.Rpc_gateway_login_token(req)
}

func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
	if accountId, ok := conn.UserPtr.(uint32); ok { //玩家断线，且没重连
		if c := logic.GetClientConn(accountId); c == nil || c.IsClose() {
			if p, ok := logic.GetGameRpc(accountId); ok { //通知游戏服
				p.CallRpcSafe(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
					buf.WriteUInt16(enum.Rpc_game_logout)
					buf.WriteUInt32(accountId)
				}, nil)
			}
			//清空缓存
			logic.DelClientConn(accountId)
			logic.DelRouteGame(accountId)
		}
	} else if ptr, ok := conn.UserPtr.(*meta.Meta); ok {
		if ptr.Module == "game" { //游戏服断开

		}
	}
}

func Rpc_gateway_relay_module(req, ack *common.NetPack, conn *tcp.TCPConn) {
	oldReqKey := req.GetReqKey()
	logic.Rpc_gateway_relay_module(req, func(backBuf *common.NetPack) {
		//异步回调，不能直接用ack
		backBuf.SetReqKey(oldReqKey)
		conn.WriteMsg(backBuf)
	})
}
func Rpc_gateway_relay_modules(req, ack *common.NetPack, conn *tcp.TCPConn) {
	oldReqKey := req.GetReqKey()
	logic.Rpc_gateway_relay_modules(req, func(backBuf *common.NetPack) {
		//异步回调，不能直接用ack
		backBuf.SetReqKey(oldReqKey)
		conn.WriteMsg(backBuf)
	})
}
func Rpc_gateway_relay_player_msg(req, _ *common.NetPack, conn *tcp.TCPConn) {
	oldReqKey := req.GetReqKey()
	logic.Rpc_gateway_relay_player_msg(req, func(backBuf *common.NetPack) {
		//异步回调，不能直接用ack
		backBuf.SetReqKey(oldReqKey)
		conn.WriteMsg(backBuf)
	})
}
