package main

import (
	"common"
	"common/net/meta"
	"conf"
	"dbmgo"
	"gamelog"
	_ "generate_out/rpc/svr_center"
	"netConfig"
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
	gamelog.SetLevel(gamelog.Lv_Debug)
	InitConf()

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_account", -1)
	dbmgo.InitWithUser(pMeta.IP, pMeta.TcpPort, pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	account.G_AccountMgr.Init()

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	component.RegisterToZookeeper()

	print("----Center Server Start-----")
	if !netConfig.CreateNetSvr(K_Module_Name, K_Module_SvrID) {
		print("----Center NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &meta.G_SvrNets,
		"conf_svr": &conf.SvrCsv,
	}
	common.LoadAllCsv()

	netConfig.G_Local_Meta = meta.GetMeta(K_Module_Name, K_Module_SvrID)
}
