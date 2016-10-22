package main

import (
	"common"
	"fmt"
	"gamelog"
	"netConfig"
	"os"
	"strconv"
	// "svr_battle/logic"
	"svr_battle/api"
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
	common.RegConsoleCmd("addsvrto", HandCmd_AddTempSvrToGame)

	//注册所有tcp消息处理方法
	RegBattleTcpMsgHandler()

	gamelog.Warn("----Battle Server Start[%d]-----", svrID)
	if netConfig.CreateNetSvr("battle", svrID) == false {
		gamelog.Error("----Battle NetSvr Failed[%d]-----", svrID)
	}
}

func HandCmd_SetLogLevel(args []string) bool {
	if len(args) < 2 {
		gamelog.Error("Lack of param")
		return false
	}
	level, err := strconv.Atoi(args[1])
	if err != nil {
		gamelog.Error("HandCmd_SetLogLevel Error : Invalid param :%s", args[1])
		return false
	}
	gamelog.SetLevel(level)
	return true
}
func HandCmd_AddTempSvrToGame(args []string) bool {
	if len(args) < 2 {
		gamelog.Error("Lack of param")
		return false
	}
	gameSvrID, err := strconv.Atoi(args[1])
	if err != nil {
		gamelog.Error("HandCmd_AddTempSvr Error : Invalid param :%s", args[1])
		return false
	}

	gameAddr := netConfig.GetAddr("game", gameSvrID)

	cfg := netConfig.G_SvrNetCfg[netConfig.G_Local_Module]
	for _, v := range cfg.Listen {
		if v.SvrID == netConfig.G_Local_SvrID {
			selfAddr := fmt.Sprintf("%s:%d", v.IP, v.TcpPort)
			api.AddTempBattleSvr(gameAddr, selfAddr, v.SvrID)
			return true
		}
	}
	return false
}
