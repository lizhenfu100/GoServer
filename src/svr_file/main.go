package main

import (
	"common/console"
	"common/file"
	"common/net/meta"
	"conf"
	"gamelog"
	_ "generate_out/rpc/svr_file"
	"netConfig"
)

const (
	Module_Name  = "file"
	Module_SvrID = 0
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(Module_Name)
	InitConf()

	//开启控制台窗口，可以接受一些调试命令
	console.StartConsole()

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

	netConfig.G_Local_Meta = meta.GetMeta(Module_Name, Module_SvrID)
}
