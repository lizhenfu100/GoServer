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
		//! Battle
		{"/battle_echo", logic.Handle_Battle_Echo},

		//! SDK
		{"/create_recharge_order", logic.Handle_Create_Recharge_Order},
		{"/sdk_recharge_success", logic.Handle_Recharge_Success},
	}

	//! register
	max := len(configSlice)
	for i := 0; i < max; i++ {
		data := &configSlice[i]
		http.HandleFunc(data.url, data.fun)
	}
}

func RegGamesvrTcpMsgHandler() {
	type TTcpFuncInfo struct {
		msgID uint16
		fun   func(*tcp.TCPConn, []byte)
	}
	var configSlice = []TTcpFuncInfo{
		{1, logic.Hand_Msg_1},
	}

	//! register
	max := len(configSlice)
	for i := 0; i < max; i++ {
		data := &configSlice[i]
		tcp.G_HandlerMsgMap[data.msgID] = data.fun
	}
}
