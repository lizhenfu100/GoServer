package api

import (
	"encoding/json"
	"http"
	"netConfig"
)

var (
	g_cache_sdk_addr string
)

// strKey = "/create_recharge_order"
func PostSdkReq(strKey string, pMsg interface{}) ([]byte, error) {
	if g_cache_sdk_addr == "" {
		g_cache_sdk_addr = netConfig.GetHttpAddr("sdk", -1)
	}

	buf, _ := json.Marshal(pMsg)
	url := g_cache_sdk_addr + strKey

	return http.PostReq(url, buf)
}
