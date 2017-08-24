package main

import (
	"common"
	"gamelog"
	"netConfig"
	"svr_file/logic"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("file")
	gamelog.SetLevel(0)

	//设置mongodb的服务器地址
	// mongodb.Init(conf.GameDbAddr)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	InitConf()

	gamelog.Warn("----File Server Start-----")
	if netConfig.CreateNetSvr("file", 0) == false {
		gamelog.Error("----File NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()

	netConfig.RegHttpHandler(map[string]netConfig.HttpHandle{
		"":       logic.Handle_File_Download,
		"upload": logic.Handle_File_Upload,
	})
	netConfig.RegHttpRpc(map[string]netConfig.HttpRpc{
		//! Client
		"rpc_update_file_list": logic.Rpc_Update_File_List,
	})
}
