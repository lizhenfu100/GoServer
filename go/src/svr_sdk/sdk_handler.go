package main

import (
	"net/http"
	"svr_sdk/logic"
)

func RegSdkMsgHandler() {
	var config = map[string]func(http.ResponseWriter, *http.Request){
		//! From Gamesvr
		"create_recharge_order": logic.HandSvr_CreateRechargeOrder,

		//! From 第三方
		"sdk_recharge_success": logic.HandSdk_RechargeSuccess,
	}
	//! register
	for k, v := range config {
		http.HandleFunc("/"+k, v)
	}
}
