package logic

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig/meta"
	"tcp"
)

func MainLoop() {
	tcp.G_RpcQueue.Loop()
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
	if accountId, ok := conn.UserPtr.(uint32); ok { //玩家离线
		//通知游戏服
		if p := GetGameConn(accountId); p != nil {
			p.CallRpc(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
				buf.WriteUInt16(enum.Rpc_game_logout)
				buf.WriteUInt32(accountId)
			}, nil)
		}
		//清空缓存
		DelClientConn(accountId)
		DelGameConn(accountId)
	} else if ptr, ok := conn.UserPtr.(*meta.Meta); ok && ptr.Module == "game" { //游戏服断开
	}
}
