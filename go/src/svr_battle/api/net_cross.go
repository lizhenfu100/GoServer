package api

import (
	"common"
	"netConfig"
	"tcp"
)

var (
	g_cache_cross_conn *tcp.TCPConn
)

func CallRpcCross(rpc string, sendFun, recvFun func(*common.NetPack)) {
	GetCrossConn().CallRpc(rpc, sendFun, recvFun)
}
func GetCrossConn() *tcp.TCPConn {
	if g_cache_cross_conn == nil || g_cache_cross_conn.IsClose() {
		g_cache_cross_conn = netConfig.GetTcpConn("cross", -1)
	}
	return g_cache_cross_conn
}
