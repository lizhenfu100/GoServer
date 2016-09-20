package api

import (
	"netConfig"
	"tcp"
)

var (
	g_cache_battle_conn *tcp.TCPConn
)

func SendToBattle(msgID uint16, msgdata []byte) {
	if g_cache_battle_conn == nil {
		g_cache_battle_conn = netConfig.GetTcpConn("battle", 0)
	}
	g_cache_battle_conn.WriteMsg(msgID, msgdata)
}
