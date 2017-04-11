package main

import (
	"common"
	"gamelog"
	"netConfig"
	// "mongodb"
	"strconv"
	"svr_game/api"
	// "svr_game/logic"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("game")
	gamelog.SetLevel(0)

	//设置mongodb的服务器地址
	// mongodb.Init(appconfig.GameDbAddr)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()
	common.RegConsoleCmd("setloglevel", HandCmd_SetLogLevel)

	//注册所有http消息处理方法
	RegGamesvrHttpMsgHandler()
	RegGamesvrTcpMsgHandler()
	RegGamesvrCsv()

	api.InitBattleSvrID()

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
