package main

import (
	"net/http"
	"svr_game/sdk"
)

func RegSdkMsgHandler() {
	var config = map[string]func(http.ResponseWriter, *http.Request){
		//! SDK
		"create_recharge_order": sdk.Handle_Create_Recharge_Order,
		"sdk_recharge_success":  sdk.Handle_Recharge_Success,
	}
	//! register
	for k, v := range config {
		http.HandleFunc("/"+k, v)
	}
}
