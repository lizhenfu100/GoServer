package player

import (
	"common"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"netConfig/meta"
)

func NotifyOnlineNum() {
	ids, _ := meta.GetModuleIDs("login", netConfig.G_Local_Meta.Version)
	for _, id := range ids {
		addr := netConfig.GetHttpAddr("login", id)
		http.CallRpc(addr, enum.Rpc_login_set_player_cnt, func(buf *common.NetPack) {
			buf.WriteInt(netConfig.G_Local_Meta.SvrID)
			buf.WriteInt32(g_player_cnt)
		}, nil)
	}
}
