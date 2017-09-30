package main

import (
	"common"
	"gamelog"
	"netConfig"

	_ "generate/rpc/sdk"
)

// 1 开一个http server
// 2 读取gamesvr list配置表，取得各游戏服的地址 —— 能够根据svrId往各游戏服推送数据
func main() {
	//初始化日志系统
	gamelog.InitLogger("sdk")
	gamelog.SetLevel(0)

	InitConf()

	gamelog.Warn("----Sdk Server Start-----")
	if !netConfig.CreateNetSvr("sdk", 0) {
		gamelog.Error("----Sdk NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()
}
