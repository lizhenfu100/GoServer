package api

import (
	"common"
	"gamelog"
	"http"
	"netConfig"
)

var (
	g_cache_game_addr = make(map[int]string)
)

// strKey = "sdk_recharge_info"
func RelayToGamesvr(svrId int, strKey string, packet *common.NetPack) {
	addr, ok := g_cache_game_addr[svrId]
	if false == ok {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr[svrId] = addr
	}

	if _, err := http.PostReq(addr+strKey, packet.DataPtr); err != nil {
		gamelog.Error("RelayToGamesvr svrId(%d) fail: %s", svrId, err.Error())
	}
}
func GetGamesvrCfgLst() (ret []*netConfig.TNetConfig) {
	ids := http.GetRegModuleIDs("game")
	for _, id := range ids {
		cfg := netConfig.GetNetCfg("game", &id)
		ret = append(ret, cfg)
	}
	return
}
