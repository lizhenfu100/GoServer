package main

import (
	"common/console"
	"common/file"
	"conf"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_gateway"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_gateway/logic"
	"shared_svr/zookeeper/component"
)

const (
	kModuleName = "gateway"
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
