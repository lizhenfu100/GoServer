package main

import (
	"common"
	"common/net/meta"
	"conf"
	"dbmgo"
	"gamelog"
	_ "generate_out/rpc/svr_game"
	"http"
	"netConfig"
	"strconv"
	"svr_game/logic"
	"svr_game/logic/player"
	"zookeeper/component"
)

const (
	K_Module_Name  = "game"
	K_Module_SvrID = 1
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(K_Module_Name)
	gamelog.SetLevel(gamelog.Lv_Debug)
	InitConf()

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_game", 1)
	dbmgo.InitWithUser(pMeta.IP, pMeta.TcpPort, pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	player.InitDB()

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()
	common.RegConsoleCmd("MakeFriends", HandCmd_MakeFriends)

	component.RegisterToZookeeper()

	go logic.MainLoop()

	print("----Game Server Start-----")
	if !netConfig.CreateNetSvr(K_Module_Name, K_Module_SvrID) {
		print("----Game NetSvr Failed-----")
	}
}
func HandCmd_MakeFriends(args []string) bool {
	pid1, err1 := strconv.Atoi(args[1])
	pid2, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		gamelog.Error("HandCmd_MakeFriends => Invalid param:%s, %s", args[1], args[2])
		return false
	}
	player1 := player.FindWithDB_PlayerId(uint32(pid1))
	player2 := player.FindWithDB_PlayerId(uint32(pid2))
	if player1 != nil && player2 != nil {
		player1.AsyncNotify(func(player *player.TPlayer) {
			player.Friend.AddFriend(player2.PlayerID, player2.Name)
		})
		player2.AsyncNotify(func(player *player.TPlayer) {
			player.Friend.AddFriend(player1.PlayerID, player1.Name)
		})
		return true
	}
	return false
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &meta.G_SvrNets,
		"conf_svr": &conf.SvrCsv,
	}
	common.LoadAllCsv()

	http.G_Before_Recv_Player = player.BeforeRecvNetMsg
	http.G_After_Recv_Player = player.AfterRecvNetMsg

	netConfig.G_Local_Meta = meta.GetMeta(K_Module_Name, K_Module_SvrID)
}
