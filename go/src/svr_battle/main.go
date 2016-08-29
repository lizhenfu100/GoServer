package main

import (
	"common"
	"gamelog"
	"netConfig"
	"strconv"
	// "svr_battle/logic"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("battle", true)
	gamelog.SetLevel(0)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()
	common.RegConsoleCmd("setloglevel", HandCmd_SetLogLevel)

	//注册所有tcp消息处理方法
	RegBattleTcpMsgHandler()

	gamelog.Warn("----Battle Server Start-----")
	if netConfig.CreateNetSvr("battle") == false {
		gamelog.Error("----Battle NetSvr Failed-----")
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
