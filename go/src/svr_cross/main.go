package main

import (
	"common"
	"gamelog"
	"netConfig"

	_ "generate_out/rpc/cross"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("cross")
	gamelog.SetLevel(0)

	InitConf()

	//设置mongodb的服务器地址
	// mongodb.Init(conf.GameDbAddr)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	gamelog.Warn("----Cross Server Start-----")
	if !netConfig.CreateNetSvr("cross", 0) {
		gamelog.Error("----Cross NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
	}
	common.LoadAllCsv()
}
