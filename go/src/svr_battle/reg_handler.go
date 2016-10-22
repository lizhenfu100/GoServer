package main

import (
	"common"
	"netConfig"
	"svr_battle/logic"
	"tcp"
)

//注册http消息处理方法
func RegBattleTcpMsgHandler() {
	var config = map[uint16]func(*tcp.TCPConn, []byte){
		1:   logic.Hand_Msg_1,
		100: logic.Hand_Login,
		101: logic.Hand_Logout,
	}

	//! register
	for k, v := range config {
		tcp.G_HandlerMsgMap[k] = v
	}
}
func RegBattleCsv() {
	var config = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
	}
	//! register
	for k, v := range config {
		common.G_CsvParserMap[k] = v
	}
}
