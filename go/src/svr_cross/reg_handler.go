package main

import (
	"common"
	"netConfig"
	"svr_cross/logic"
	"tcp"
)

func RegCrossTcpMsgHandler() {
	var config = map[uint16]func(*tcp.TCPConn, *common.NetPack){
		0: logic.Rpc_Echo,
	}

	//! register
	for k, v := range config {
		tcp.G_HandlerMsgMap[k] = v
	}
}
func RegCrossCsv() {
	var config = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
	}
	//! register
	for k, v := range config {
		common.G_CsvParserMap[k] = v
	}
	common.LoadAllCsv()
}
