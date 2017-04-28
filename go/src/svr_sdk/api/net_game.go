package api

import (
	"encoding/json"
	"http"
	"netConfig"
)

var (
	g_cache_game_addr = make(map[int]string)
)

// strKey = "sdk_recharge_info"
func SendToGame(svrId int, strKey string, pMsg interface{}) []byte {
	addr, ok := g_cache_game_addr[svrId]
	if false == ok {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr[svrId] = addr
	}

	data, _ := json.Marshal(pMsg)
	return http.PostReq(addr+strKey, data)
}
