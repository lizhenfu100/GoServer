package main

import (
	"common"
	"dbmgo"
	"gamelog"
	"netConfig"

	_ "generate_out/rpc/center"
	"svr_center/logic/account"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("center")
	gamelog.SetLevel(0)

	InitConf()

	//设置mongodb的服务器地址
	var id int
	cfg := netConfig.GetNetCfg("db_account", &id)
	dbmgo.Init(cfg.IP, cfg.TcpPort, cfg.SvrName)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	account.G_AccountMgr.Init()

	gamelog.Warn("----Center Server Start-----")
	if !netConfig.CreateNetSvr("center", 0) {
		gamelog.Error("----Center NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
	}
	common.LoadAllCsv()
}
