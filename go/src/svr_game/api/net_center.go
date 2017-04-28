package api

import (
	"common"
	"http"
	"netConfig"
)

var (
	g_cache_center_addr string
)

// strKey = "create_recharge_order"
func SendToCenter(strKey string, packet *common.NetPack) []byte {
	if g_cache_center_addr == "" {
		g_cache_center_addr = netConfig.GetHttpAddr("center", -1)
	}
	return http.PostReq(g_cache_center_addr+strKey, packet.DataPtr)
}
