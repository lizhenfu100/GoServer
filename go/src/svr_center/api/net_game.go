package api

import (
	"http"
	"netConfig"
)

var (
	g_cache_game_addr = make(map[int]string)
)

// strKey = "sdk_recharge_info"
func SendToGame(svrId int, strKey string, buf []byte) []byte {
	addr, ok := g_cache_game_addr[svrId]
	if false == ok {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr[svrId] = addr
	}
	return http.PostReq(addr+strKey, buf)
}
func GetRegGamesvrCfgLst() (ret []*netConfig.TNetConfig) {
	ids := http.GetRegModuleIDs("game")
	for _, id := range ids {
		cfg := netConfig.GetNetCfg("game", &id)
		ret = append(ret, cfg)
	}
	return
}
