package main

import (
	"common"
	"net/http"
	"netConfig"
	"svr_game/cross"
	"svr_game/logic"
	"svr_game/sdk"
	"tcp"
)

//注册http消息处理方法
func RegGamesvrHttpMsgHandler() {
	var config = map[string]func(http.ResponseWriter, *http.Request){
		//! Client
		"/battle_echo": logic.Handle_Client2Battle_Echo,

		//! SDK
		"/create_recharge_order": sdk.Handle_Create_Recharge_Order,
		"/sdk_recharge_success":  sdk.Handle_Recharge_Success,
	}

	//! register
	for k, v := range config {
		http.HandleFunc(k, v)
	}
}
func RegGamesvrTcpMsgHandler() {
	var config = map[uint16]func(*tcp.TCPConn, *common.NetPack){
		0: cross.Rpc_Echo,
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
