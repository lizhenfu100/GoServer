package main

import (
	"common/console"
	"common/file"
	"common/tool/email"
	"conf"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_center"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_center/logic"
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

	netConfig.RunNetSvr(false)
	logic.MainLoop()
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.SvrCsv)
	file.RegCsvType("csv/email/email.csv", email.EmailCsv)
	file.RegCsvType("csv/email/invalid.csv", email.InvalidCsv)
	file.LoadAllCsv()
	console.Init()
}
