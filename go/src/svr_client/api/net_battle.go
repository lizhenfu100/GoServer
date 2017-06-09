package api

import (
	"common"
	"netConfig"
	"tcp"
)

var (
	g_cache_battle_conn = make(map[int]*tcp.TCPConn)
)

func SendToBattle(svrID int, msg *common.NetPack) {
	conn, _ := g_cache_battle_conn[svrID]
	if conn == nil || conn.IsClose() {
		conn = netConfig.GetTcpConn("battle", svrID)
		g_cache_battle_conn[svrID] = conn
	}
	conn.WriteMsg(msg)
}
