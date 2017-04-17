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
func RegHttpMsgHandler() {
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
func RegTcpMsgHandler() {
	var config = map[string]func(*tcp.TCPConn, *common.NetPack){
		"rpc_echo": cross.Rpc_Echo,
	}
	//! register
	for k, v := range config {
		tcp.G_HandlerMsgMap[common.RpcToOpcode(k)] = v
	}
}
func RegCsv() {
	var config = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	//! register
	for k, v := range config {
		common.G_CsvParserMap[k] = v
	}
	common.LoadAllCsv()
}
