package main

import (
	"common"
	//"fmt"
	"gamelog"
	"netConfig"

	"svr_battle/logic"
)

func main() {
	//获取命令参数 svrID
	// svrID, err := strconv.Atoi(os.Args[1])
	// if err != nil {
	// 	gamelog.Error("Please input the svrID!!!")
	// 	return
	// }
	svrID := 1

	//初始化日志系统
	gamelog.InitLogger("battle")
	gamelog.SetLevel(0)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	InitConf()

	gamelog.Warn("----Battle Server Start[%d]-----", svrID)
	if netConfig.CreateNetSvr("battle", svrID) == false {
		gamelog.Error("----Battle NetSvr Failed[%d]-----", svrID)
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()

	netConfig.RegTcpRpc(map[string]netConfig.TcpHandle{
		"rpc_echo": logic.Rpc_Echo,
	})
}
