package main

import (
	"common/console"
	"common/file"
	"conf"
	"dbmgo"
	"gamelog"
	_ "generate_out/rpc/svr_game"
	"netConfig"
	"netConfig/meta"
	"shared_svr/zookeeper/component"
	"svr_game/logic"
	"svr_game/logic/player"
)

const (
	K_Module_Name  = "game"
	K_Module_SvrID = 1
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(K_Module_Name)
	InitConf()

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_game", K_Module_SvrID)
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
	file.G_Csv_Map = map[string]interface{}{
		"conf_net": &metaCfg,
		"conf_svr": &conf.SvrCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	netConfig.G_Local_Meta = meta.GetMeta(K_Module_Name, K_Module_SvrID)
}
