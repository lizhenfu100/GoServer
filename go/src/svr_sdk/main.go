package main

import (
	"common"
	"gamelog"
	"netConfig"
	"strconv"
	"svr_sdk/logic"
)

// 1 开一个http server
// 2 读取gamesvr list配置表，取得各游戏服的地址 —— 能够根据svrId往各游戏服推送数据
func main() {
	//初始化日志系统
	gamelog.InitLogger("sdk")
	gamelog.SetLevel(0)

	//设置mongodb的服务器地址
	// mongodb.Init(conf.GameDbAddr)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()
	common.RegConsoleCmd("setloglevel", HandCmd_SetLogLevel)

	InitConf()

	gamelog.Warn("----Sdk Server Start-----")
	if netConfig.CreateNetSvr("sdk", 0) == false {
		gamelog.Error("----Sdk NetSvr Failed-----")
	}
}
func HandCmd_SetLogLevel(args []string) bool {
	level, err := strconv.Atoi(args[1])
	if err != nil {
		gamelog.Error("HandCmd_SetLogLevel Error : Invalid param :%s", args[1])
		return true
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

	netConfig.RegHttpHandler(map[string]netConfig.HttpHandle{
		//! From Gamesvr
		"create_recharge_order": logic.HandSvr_CreateRechargeOrder,

		//! From 第三方
		"sdk_recharge_success": logic.HandSdk_RechargeSuccess,
	})
}
