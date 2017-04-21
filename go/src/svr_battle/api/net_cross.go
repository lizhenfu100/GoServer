package api

import (
	"common"
	"netConfig"
	"tcp"
)

var (
	g_cache_cross_conn *tcp.TCPConn
)

func SendToCross(msg *common.NetPack) {
	if g_cache_cross_conn == nil {
		g_cache_cross_conn = netConfig.GetTcpConn("cross", -1)
	}
	g_cache_cross_conn.WriteMsg(msg)
}
