package api

import (
	"netConfig"
	"tcp"
)

var (
	g_cache_game_conn *tcp.TCPConn
)

func SendToGame(msgID uint16, msgdata []byte) {
	if g_cache_game_conn == nil {
		g_cache_game_conn = netConfig.GetTcpConn("game", 0)
	}
	g_cache_game_conn.WriteMsg(msgID, msgdata)
}
