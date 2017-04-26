package api

import (
	"encoding/json"
	"gamelog"
	"http"
	"netConfig"
)

var (
	g_cache_game_addr = make(map[int]string)
)

// strKey = "sdk_recharge_info"
func RelayToGamesvr(svrId int, strKey string, pMsg interface{}) {
	addr, ok := g_cache_game_addr[svrId]
	if false == ok {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr[svrId] = addr
	}

	data, _ := json.Marshal(pMsg)
	if http.PostReq(addr+strKey, data) == nil {
		gamelog.Error("RelayToGamesvr svrId(%d) fail", svrId)
	}
}
