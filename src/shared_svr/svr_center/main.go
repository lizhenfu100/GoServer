package main

import (
	"common/console"
	"common/console/shutdown"
	"common/file"
	"conf"
	"dbmgo"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_center"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_center/logic"
	"shared_svr/zookeeper/component"
)

const kModuleName = "center"

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(kModuleName, svrId)

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_account", svrId)
	dbmgo.InitWithUser(pMeta.IP, pMeta.Port(), pMeta.SvrName,
		conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)

	component.RegisterToZookeeper()

	go netConfig.RunNetSvr()
	logic.MainLoop()
}
func InitConf() {
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"csv/conf_net.csv": &metaCfg,
		"csv/conf_svr.csv": &conf.SvrCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	console.Init()
	console.RegShutdown(shutdown.Default)
}
