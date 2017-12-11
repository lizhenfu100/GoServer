package api

import (
	"encoding/json"
	"http"
	"netConfig"
)

var (
	g_cache_game_addr = make(map[int]string)
)

// ------------------------------------------------------------
//! game
func SendToGame(svrId int, strKey string, pMsg interface{}) []byte { // strKey = "sdk_recharge_info"
	addr, ok := g_cache_game_addr[svrId]
	if false == ok {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr[svrId] = addr
	}

	data, _ := json.Marshal(pMsg)
	return http.PostReq(addr+strKey, data)
}
