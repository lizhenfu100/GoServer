package api

import (
	"common"
	"http"
	"netConfig"
)

var (
	g_cache_center_addr string
)

func CallRpcCenter(rpc string, sendFun, recvFun func(*common.NetPack)) {
	if g_cache_center_addr == "" {
		g_cache_center_addr = netConfig.GetHttpAddr("center", -1)
	}
	http.CallRpc(g_cache_center_addr, rpc, sendFun, recvFun)
}
