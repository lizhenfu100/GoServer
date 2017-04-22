package main

import (
	"common"
	"conf"
	"dbmgo"
	"gamelog"
	"netConfig"
	"strconv"

	"svr_game/cross"
	"svr_game/logic"
	"svr_game/logic/player"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("game")
	gamelog.SetLevel(0)

	//设置mongodb的服务器地址
	dbmgo.Init(conf.GameDbAddr, conf.GameDbName)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()
	common.RegConsoleCmd("setloglevel", HandCmd_SetLogLevel)

	InitConf()
	common.LoadAllCsv()
	netConfig.RegMsgHandler()
	RegSdkMsgHandler()

	gamelog.Warn("----Game Server Start-----")
	if netConfig.CreateNetSvr("game", 1) == false {
		gamelog.Error("----Game NetSvr Failed-----")
	}
}
func HandCmd_SetLogLevel(args []string) bool {
	level, err := strconv.Atoi(args[1])
	if err != nil {
		gamelog.Error("HandCmd_SetLogLevel Error : Invalid param :%s", args[1])
		return true
	}
	gamelog.SetLevel(level)
	return true
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	netConfig.G_Tcp_Handler = map[string]netConfig.TcpHandle{
		"rpc_echo": cross.Rpc_Echo,
	}
	netConfig.G_Http_Handler = map[string]netConfig.HttpHandle{
		//! Client
		"battle_echo":       logic.Rpc_Client2Battle_Echo,
		"rpc_test_mongodb":  logic.Rpc_test_mongodb,
		"rpc_player_login":  player.Rpc_Player_Login,
		"rpc_player_logout": player.Rpc_Player_Logout,
		"rpc_player_create": player.Rpc_Player_Create,
	}
}
