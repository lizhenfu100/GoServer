package api

import (
	"common"
	"http"
	"netConfig"
)

var (
	g_cache_login_addr string
)

func CallRpcLogin(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if g_cache_login_addr == "" {
		g_cache_login_addr = netConfig.GetHttpAddr("login", -1)
	}
	http.CallRpc(g_cache_login_addr, rid, sendFun, recvFun)
}
