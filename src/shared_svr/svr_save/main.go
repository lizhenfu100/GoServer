package main

import (
	"common/console"
	"common/console/shutdown"
	"common/file"
	"common/tool/email"
	"conf"
	"dbmgo"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_save"
	"netConfig"
	"netConfig/meta"
	conf2 "shared_svr/svr_save/conf"
)

const kModuleName = "save"

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
	pMeta := meta.GetMeta("db_save", svrId)
	dbmgo.InitWithUser(pMeta.IP, pMeta.Port(), pMeta.SvrName,
		conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)

	netConfig.RunNetSvr()
}
func InitConf() {
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"csv/conf_net.csv":    &metaCfg,
		"csv/conf_svr.csv":    &conf.SvrCsv,
		"csv/email/email.csv": &email.G_Email,
		"csv/save/const.csv":  &conf2.Const,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	console.Init()
	console.RegShutdown(shutdown.Default)
}
