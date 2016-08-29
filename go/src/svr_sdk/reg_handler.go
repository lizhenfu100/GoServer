package main

import (
	"net/http"
	"svr_sdk/logic"
)

//注册http消息处理方法
func RegSdkHttpMsgHandler() {
	type THttpFuncInfo struct {
		url string
		fun func(http.ResponseWriter, *http.Request)
	}
	var configSlice = []THttpFuncInfo{
		//! From Gamesvr
		{"/create_recharge_order", logic.HandSvr_CreateRechargeOrder},
		{"/reg_gamesvr_addr", logic.HandSvr_GamesvrAddr},

		//! From 第三方
		{"/sdk_recharge_success", logic.HandSdk_RechargeSuccess},
	}

	max := len(configSlice)
	for i := 0; i < max; i++ {
		data := &configSlice[i]
		http.HandleFunc(data.url, data.fun)
	}
}
