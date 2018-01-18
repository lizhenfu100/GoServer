package main

import (
	"common"
	"common/console"
	"common/net/meta"
	"conf"
	"dbmgo"
	"gamelog"
	_ "generate_out/rpc/svr_center"
	"netConfig"
	"svr_center/logic"
	"svr_center/logic/account"
	"zookeeper/component"
)

const (
	K_Module_Name  = "center"
	K_Module_SvrID = 0
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(K_Module_Name)
	if conf.IsDebug {
		gamelog.SetLevel(gamelog.Lv_Debug)
	} else {
		gamelog.SetLevel(gamelog.Lv_Info)
	}
	InitConf()

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_account", 0)
	dbmgo.InitWithUser(pMeta.IP, pMeta.TcpPort, pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	account.G_AccountMgr.Init()

	//开启控制台窗口，可以接受一些调试命令
	console.StartConsole()

	component.RegisterToZookeeper()

	go logic.MainLoop()

	print("----Center Server Start-----")
	if !netConfig.CreateNetSvr(K_Module_Name, K_Module_SvrID) {
		print("----Center NetSvr Failed-----")
	}
}
func InitConf() {
	var metaCfg []meta.Meta
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &metaCfg,
		"conf_svr": &conf.SvrCsv,
	}
	common.LoadAllCsv()
	meta.InitConf(metaCfg)

	netConfig.G_Local_Meta = meta.GetMeta(K_Module_Name, K_Module_SvrID)
}
