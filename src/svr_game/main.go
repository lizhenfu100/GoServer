package main

import (
	"common/console"
	"common/file"
	"conf"
	"dbmgo"
	"flag"
	"gamelog"
	_ "generate_out/rpc/svr_game"
	"netConfig"
	"netConfig/meta"
	"shared_svr/zookeeper/component"
	"svr_game/logic"
	"svr_game/logic/player"
)

const (
	kModuleName = "game"
)

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	netConfig.G_Local_Meta = meta.GetMeta(kModuleName, svrId)

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_game", svrId)
	dbmgo.InitWithUser(pMeta.IP, pMeta.Port(), pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	player.InitDB()

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
	console.Init()
}
