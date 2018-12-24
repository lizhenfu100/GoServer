package main

import (
	"common/console"
	"common/file"
	"conf"
	"dbmgo"
	"gamelog"
	_ "generate_out/rpc/svr_sdk"
	"netConfig"
	"netConfig/meta"
	"svr_sdk/msg"
)

const (
	kModuleName = "sdk"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(kModuleName, 0)

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_sdk", 0)
	dbmgo.InitWithUser(pMeta.IP, pMeta.Port(), pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	msg.InitDB()

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
