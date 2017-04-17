package main

import (
	"common"
	"net/http"
	"netConfig"
	"svr_center/logic"
)

func RegCenterHttpMsgHandler() {
	var config = map[string]func(http.ResponseWriter, *http.Request){
		//! From Gamesvr

		//! From Client
		"/rpc_get_gamesvr_lst": logic.Rpc_GetGameSvrLst,
	}

	//! register
	for k, v := range config {
		http.HandleFunc(k, v)
	}
}
func RegCenterCsv() {
	var config = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
	}
	//! register
	for k, v := range config {
		common.G_CsvParserMap[k] = v
	}
	common.LoadAllCsv()
}
