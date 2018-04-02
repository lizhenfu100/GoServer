package main

import (
	"common/console"
	"common/file"
	"conf"
	"dbmgo"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_center"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_center/logic"
	"shared_svr/svr_center/account"
	"shared_svr/zookeeper/component"
)

const (
	K_Module_Name  = "center"
	K_Module_SvrID = 1
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(K_Module_Name)
	InitConf()

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_account", 0)
	dbmgo.InitWithUser(pMeta.IP, pMeta.Port(), pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	account.InitDB()

	//开启控制台窗口，可以接受一些调试命令
	console.StartConsole()

	component.RegisterToZookeeper()

	go logic.MainLoop()

	netConfig.RunNetSvr()
}
func InitConf() {
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"conf_net": &metaCfg,
		"conf_svr": &conf.SvrCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	netConfig.G_Local_Meta = meta.GetMeta(K_Module_Name, K_Module_SvrID)
}
