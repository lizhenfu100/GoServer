package api

import (
	"common"
	"netConfig"
	"tcp"
)

//Notice：TCPConn是对真正net.Conn的包装，发生断线重连时，会执行tcp.TCPConn.ResetConn()，所以外部缓存的tcp.TCPConn仍有效，无需更新
var (
	g_cache_cross_conn *tcp.TCPConn
)

func CallRpcCross(rpc string, sendFun, recvFun func(*common.NetPack)) {
	if g_cache_cross_conn == nil || g_cache_cross_conn.IsClose() {
		g_cache_cross_conn = netConfig.GetTcpConn("cross", -1)
	}
	g_cache_cross_conn.CallRpc(rpc, sendFun, recvFun)
}
