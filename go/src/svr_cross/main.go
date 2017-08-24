package main

import (
	"common"
	"gamelog"
	"netConfig"

	"svr_cross/logic"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("cross")
	gamelog.SetLevel(0)

	//设置mongodb的服务器地址
	// mongodb.Init(conf.GameDbAddr)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	InitConf()

	gamelog.Warn("----Cross Server Start-----")
	if netConfig.CreateNetSvr("cross", 0) == false {
		gamelog.Error("----Cross NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()

	netConfig.RegTcpRpc(map[string]netConfig.TcpHandle{
		"rpc_echo":                    logic.Rpc_Echo,
		"rpc_cross_relay_battle_data": logic.Rpc_Relay_Battle_Data,
	})
}
