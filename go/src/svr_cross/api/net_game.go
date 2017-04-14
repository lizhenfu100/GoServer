package api

import (
	"common"
	"netConfig"
	"tcp"
)

var (
	g_cache_game_conn *tcp.TCPConn
)

func SendToGame(msg *common.NetPack) {
	if g_cache_game_conn == nil {
		g_cache_game_conn = netConfig.GetTcpConn("game", -1)
	}
	g_cache_game_conn.WriteMsg(msg)
}
