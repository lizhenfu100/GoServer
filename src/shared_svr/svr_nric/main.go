package main

import (
	"common/console"
	"common/file"
	"conf"
	"flag"
	"gamelog"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_nric/NRIC"
	"shared_svr/zookeeper/component"
)

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(NRIC.KDBTable)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(NRIC.KDBTable, svrId)
	
	component.RegisterToZookeeper()

	netConfig.RunNetSvr(true)
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.NilSvrCsv())
	file.LoadAllCsv()
	console.Init()
}
