package main

import (
	"common/console"
	"common/file"
	"conf"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_gateway"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_gateway/logic"
	"shared_svr/zookeeper/component"
)

const (
	K_Module_Name  = "gateway"
	K_Module_SvrID = 1
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(K_Module_Name)
	InitConf()

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
