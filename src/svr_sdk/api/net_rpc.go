package api

import (
	"encoding/json"
	"http"
	"netConfig"
	"sync"
)

var (
	g_cache_game_addr sync.Map // make(map[int]string)
)

// ------------------------------------------------------------
//! game
func SendToGame(svrId int, strKey string, pMsg interface{}) []byte { // strKey = "sdk_recharge_info"
	var addr string
	if v, ok := g_cache_game_addr.Load(svrId); ok {
		addr = v.(string)
	} else {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr.Store(svrId, addr)
	}
	data, _ := json.Marshal(pMsg)
	return http.PostReq(addr+strKey, data)
}
