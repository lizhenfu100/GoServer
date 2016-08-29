package main

import (
	"net/http"
	"svr_game/logic"
	"tcp"
)

//注册http消息处理方法
func RegGamesvrHttpMsgHandler() {
	type THttpFuncInfo struct {
		url string
		fun func(http.ResponseWriter, *http.Request)
	}
	var configSlice = []THttpFuncInfo{

		{"/battle_echo", logic.Handle_Battle_Echo},

		//! SDK
		{"/create_recharge_order", logic.Handle_Create_Recharge_Order},
		{"/sdk_recharge_success", logic.Handle_Recharge_Success},
	}

	max := len(configSlice)
	for i := 0; i < max; i++ {
		data := &configSlice[i]
		http.HandleFunc(data.url, data.fun)
	}
}

func RegGamesvrTcpMsgHandler() {
	tcp.G_HandlerMsgMap[1] = logic.Hand_Msg_1
}
