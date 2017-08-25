package main

import (
	"common"
	"conf"
	"fmt"
	"gamelog"
	"netConfig"
	"svr_file/logic"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("file")
	gamelog.SetLevel(0)

	InitConf()

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	fmt.Println("----File Server Start-----")
	if netConfig.CreateNetSvr("file", 0) == false {
		gamelog.Error("----File NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"conf_svr": &conf.SvrCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()

	for k, v := range netConfig.G_SvrNetCfg {
		fmt.Println(k, v)
	}
	fmt.Println(conf.SvrCfg)

	netConfig.RegHttpHandler(map[string]netConfig.HttpHandle{
		"":       logic.Handle_File_Download,
		"upload": logic.Handle_File_Upload,
	})
	netConfig.RegHttpRpc(map[string]netConfig.HttpRpc{
		//! Client
		"rpc_update_file_list": logic.Rpc_Update_File_List,
	})
}
