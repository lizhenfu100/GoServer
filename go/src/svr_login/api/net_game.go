package api

import (
	"common"
	"http"
	"netConfig"
)

var (
	g_cache_game_addr = make(map[int]string)
)

// rpc = "sdk_recharge_info"
func CallRpcGame(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	addr, ok := g_cache_game_addr[svrId]
	if false == ok {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr[svrId] = addr
	}
	http.CallRpc(addr, rid, sendFun, recvFun)
}
func GetRegGamesvrCfgLst() (ret []*netConfig.TNetConfig) {
	ids := netConfig.GetRegModuleIDs("game")
	for _, id := range ids {
		cfg := netConfig.GetNetCfg("game", &id)
		ret = append(ret, cfg)
	}
	return
}
