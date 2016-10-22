package main

import (
	"common"
	"net/http"
	"netConfig"
	"svr_sdk/logic"
)

//注册http消息处理方法
func RegSdkHttpMsgHandler() {
	var config = map[string]func(http.ResponseWriter, *http.Request){
		//! From Gamesvr
		"/create_recharge_order": logic.HandSvr_CreateRechargeOrder,

		//! From 第三方
		"/sdk_recharge_success": logic.HandSdk_RechargeSuccess,
	}

	//! register
	for k, v := range config {
		http.HandleFunc(k, v)
	}
}
func RegSdkCsv() {
	var config = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
	}
	//! register
	for k, v := range config {
		common.G_CsvParserMap[k] = v
	}
}
