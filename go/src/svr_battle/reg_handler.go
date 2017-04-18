package main

import (
	"common"
	"netConfig"
	"svr_battle/logic"
	"tcp"
)

func RegTcpMsgHandler() {
	var config = map[string]func(*tcp.TCPConn, *common.NetPack){
		"rpc_echo":   logic.Rpc_Echo,
		"rpc_login":  logic.Rpc_Login,
		"rpc_logout": logic.Rpc_Logout,
	}
	//! register
	for k, v := range config {
		tcp.G_HandlerMsgMap[common.RpcNameToId(k)] = v
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
