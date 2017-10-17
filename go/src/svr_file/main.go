package main

import (
	"common"
	"conf"
	"fmt"
	"gamelog"
	"netConfig"
	"svr_file/logic"

	_ "generate_out/rpc/file"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("file")
	gamelog.SetLevel(0)

	InitConf()

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	fmt.Println("----File Server Start-----")
	if !netConfig.CreateNetSvr("file", 0) {
		gamelog.Error("----File NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"conf_svr": &conf.SvrCsv,
	}
	common.LoadAllCsv()

	for k, v := range netConfig.G_SvrNetCfg {
		fmt.Println(k, v)
	}
	fmt.Println("SvrCsv: ", conf.SvrCsv)

	netConfig.RegHttpHandler(map[string]netConfig.HttpHandle{
		"":       logic.Rpc_file_download,
		"upload": logic.Rpc_file_upload,
	})
}
