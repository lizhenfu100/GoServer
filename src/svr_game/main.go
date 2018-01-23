package main

import (
	"common"
	"common/console"
	"common/net/meta"
	"conf"
	"dbmgo"
	"gamelog"
	_ "generate_out/rpc/svr_game"
	"http"
	"netConfig"
	"svr_game/logic"
	"svr_game/logic/player"
	"zookeeper/component"
)

const (
	K_Module_Name  = "game"
	K_Module_SvrID = 1
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(K_Module_Name)
	if conf.IsDebug {
		gamelog.SetLevel(gamelog.Lv_Debug)
	} else {
		gamelog.SetLevel(gamelog.Lv_Info)
		go gamelog.AutoChangeFile(K_Module_Name)
	}
	InitConf()

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_game", 1)
	dbmgo.InitWithUser(pMeta.IP, pMeta.Port(), pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	player.InitDB()

	//开启控制台窗口，可以接受一些调试命令
	console.StartConsole()

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

	http.G_Before_Recv_Player = player.BeforeRecvNetMsg
	http.G_After_Recv_Player = player.AfterRecvNetMsg

	netConfig.G_Local_Meta = meta.GetMeta(K_Module_Name, K_Module_SvrID)
}
