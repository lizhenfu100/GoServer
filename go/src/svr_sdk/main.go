package main

import (
	"common"
	"dbmgo"
	"gamelog"
	"netConfig"
	"svr_sdk/logic"

	_ "generate_out/rpc/sdk"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("sdk")
	gamelog.SetLevel(0)

	InitConf()

	//设置mongodb的服务器地址
	var id int
	cfg := netConfig.GetNetCfg("db_sdk", &id)
	dbmgo.Init(cfg.IP, cfg.TcpPort, cfg.SvrName)

	logic.InitDB()

	gamelog.Warn("----Sdk Server Start-----")
	if !netConfig.CreateNetSvr("sdk", 0) {
		gamelog.Error("----Sdk NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
	}
	common.LoadAllCsv()
}
