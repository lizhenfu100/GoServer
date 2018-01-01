package main

import (
	"common"
	"common/console"
	"common/net/meta"
	"conf"
	"gamelog"
	_ "generate_out/rpc/svr_cross"
	"netConfig"
	"svr_cross/logic"
	"zookeeper/component"
)

const (
	Module_Name  = "cross"
	Module_SvrID = 0
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(Module_Name)
	if conf.IsDebug {
		gamelog.SetLevel(gamelog.Lv_Debug)
	} else {
		gamelog.SetLevel(gamelog.Lv_Info)
	}
	InitConf()

	//开启控制台窗口，可以接受一些调试命令
	console.StartConsole()

	component.RegisterToZookeeper()

	go logic.MainLoop()

	print("----Cross Server Start-----")
	if !netConfig.CreateNetSvr(Module_Name, Module_SvrID) {
		print("----Cross NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &meta.G_SvrNets,
		"conf_svr": &conf.SvrCsv,
	}
	common.LoadAllCsv()

	netConfig.G_Local_Meta = meta.GetMeta(Module_Name, Module_SvrID)
}
