package logic

import (
	"common"
	"netConfig/meta"
	"nets/tcp"
)

func MainLoop() {
	tcp.G_RpcQueue.Loop()
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
	if ptr, ok := conn.UserPtr.(*meta.Meta); ok && ptr.Module == "battle" {
		delete(g_battle_player_cnt, ptr.SvrID)
	}
}
