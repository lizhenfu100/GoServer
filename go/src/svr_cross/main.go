package main

import (
	"common"
	"gamelog"
	"netConfig"
	"strconv"

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
	common.RegConsoleCmd("setloglevel", HandCmd_SetLogLevel)

	InitConf()

	gamelog.Warn("----Cross Server Start-----")
	if netConfig.CreateNetSvr("cross", 0) == false {
		gamelog.Error("----Cross NetSvr Failed-----")
	}
}
func HandCmd_SetLogLevel(args []string) bool {
	if len(args) < 2 {
		gamelog.Error("Lack of param")
		return false
	}
	level, err := strconv.Atoi(args[1])
	if err != nil {
		gamelog.Error("HandCmd_SetLogLevel Error : Invalid param :%s", args[1])
		return false
	}
	gamelog.SetLevel(level)
	return true
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()

	netConfig.RegTcpHandler(map[string]netConfig.TcpHandle{
		"rpc_echo":          logic.Rpc_Echo,
		"relay_battle_data": logic.Relay_Battle_Data,
	})
}
