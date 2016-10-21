package main

import (
	"common"
	"gamelog"
	"netConfig"
	"os"
	"strconv"
	// "svr_battle/logic"
)

func main() {
	//获取命令参数 svrID
	svrID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		gamelog.Error("Please input the svrID!!!")
		return
	}

	//初始化日志系统
	gamelog.InitLogger("battle")
	gamelog.SetLevel(0)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()
	common.RegConsoleCmd("setloglevel", HandCmd_SetLogLevel)

	//注册所有tcp消息处理方法
	RegBattleTcpMsgHandler()

	gamelog.Warn("----Battle Server Start-----")
	if netConfig.CreateNetSvr("battle", svrID) == false {
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
