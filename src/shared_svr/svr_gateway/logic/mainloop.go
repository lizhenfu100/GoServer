package logic

import (
	"common"
	"netConfig/meta"
	"tcp"
)

func MainLoop() {
	tcp.G_RpcQueue.Loop()
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
	if accountId, ok := conn.UserPtr.(uint32); ok {
		DelClientConn(accountId)
		DelGameConn(accountId)
	} else if ptr, ok := conn.UserPtr.(*meta.Meta); ok && ptr.Module == "game" {
		delete(g_game_player_cnt, ptr.SvrID)
	}
}
