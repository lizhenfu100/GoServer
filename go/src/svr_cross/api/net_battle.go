package api

import (
	"common"
	"netConfig"
	"tcp"
)

//Notice：TCPConn是对真正net.Conn的包装，发生断线重连时，会执行tcp.TCPConn.ResetConn()，所以外部缓存的tcp.TCPConn仍有效，无需更新
var (
	g_cache_battle_conn = make(map[int]*tcp.TCPConn)
)

func SendToBattle(svrID int, msg *common.NetPack) {
	conn, _ := g_cache_battle_conn[svrID]
	if conn == nil {
		conn = netConfig.GetTcpConn("battle", svrID)
		g_cache_battle_conn[svrID] = conn
	}
	conn.WriteMsg(msg)
}
func AllocBattleSvrID(userid int) int {
	svrIDs := tcp.GetRegModuleIDs("battle")
	idx := userid % len(svrIDs)
	return svrIDs[idx]
}
