package tcp

import (
	"common"
	"generate_out/err"
	"netConfig/meta"
	"nets/tcp"
	"shared_svr/svr_gateway/logic"
)

func Rpc_check_identity(req, ack *common.NetPack, client *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()
	if logic.CheckToken(accountId, token) {
		if p := logic.GetClientConn(accountId); p != nil && p != client {
			p.Close() //防串号
		}
		client.SetUser(accountId)
		logic.AddClientConn(accountId, client)
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Token_verify_err)
	}
}
func Rpc_set_identity(req, ack *common.NetPack, conn *tcp.TCPConn) { logic.Rpc_set_identity(req) }
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
	if accountId, ok := conn.GetUser().(uint32); ok { //玩家断线，且没重连
		if logic.TryDelClientConn(accountId) {
			logic.DelRouteGame(accountId)
		}
	} else if ptr, ok := conn.GetUser().(*meta.Meta); ok {
		if ptr.Module == "game" { //游戏服断开

		}
	}
}
func Rpc_gateway_relay(req, _ *common.NetPack, conn *tcp.TCPConn) {
	oldReqKey := req.GetReqKey()
	logic.Rpc_gateway_relay(req, func(backBuf *common.NetPack) {
		backBuf.SetReqKey(oldReqKey)
		conn.WriteMsg(backBuf) //异步回调，不能直接用ack
	})
}
