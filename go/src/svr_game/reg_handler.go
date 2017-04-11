package main

import (
	"common"
	"net/http"
	"netConfig"
	"svr_game/logic"
	"tcp"
)

//注册http消息处理方法
func RegGamesvrHttpMsgHandler() {
	var config = map[string]func(http.ResponseWriter, *http.Request){
		//! Battle
		"/battle_echo": logic.Handle_Battle_Echo,

		//! SDK
		"/create_recharge_order": logic.Handle_Create_Recharge_Order,
		"/sdk_recharge_success":  logic.Handle_Recharge_Success,

		"/add_temp_svr": logic.Handle_Add_Temp_Svr,
	}

	//! register
	for k, v := range config {
		http.HandleFunc(k, v)
	}
}
func RegGamesvrTcpMsgHandler() {
	var config = map[uint16]func(*tcp.TCPConn, []byte){
		1: logic.Hand_Msg_1,
	}

	//! register
	for k, v := range config {
		tcp.G_HandlerMsgMap[k] = v
	}
}
func RegGamesvrCsv() {
	var config = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
	}
	//! register
	for k, v := range config {
		common.G_CsvParserMap[k] = v
	}
	common.LoadAllCsv()
}
