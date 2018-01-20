package main

import (
	"common"
	"common/net/meta"
	"conf"
	"dbmgo"
	"gamelog"
	_ "generate_out/rpc/svr_sdk"
	"netConfig"
	"svr_sdk/logic"
	"svr_sdk/msg"
	"zookeeper/component"
)

const (
	K_Module_Name  = "sdk"
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
	pMeta := meta.GetMeta("db_sdk", 0)
	dbmgo.InitWithUser(pMeta.IP, pMeta.TcpPort, pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	msg.InitDB()

	component.RegisterToZookeeper()

	go logic.MainLoop()

	netConfig.RunNetSvr()
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
