package main

import (
	"common"
	"dbmgo"
	"gamelog"
	"http"
	"netConfig"
	"strconv"

	_ "generate/rpc/game"
	"svr_game/logic"
	"svr_game/logic/player"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("game")
	gamelog.SetLevel(0)

	InitConf()

	//设置mongodb的服务器地址
	var id int
	cfg := netConfig.GetNetCfg("db_game", &id)
	dbmgo.Init(cfg.IP, cfg.TcpPort, cfg.SvrName)

	player.LoadActivePlayerFormDB()

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()
	common.RegConsoleCmd("MakeFriends", HandCmd_MakeFriends)

	go logic.MainLoop()

	gamelog.Warn("----Game Server Start-----")
	if !netConfig.CreateNetSvr("game", 1) {
		gamelog.Error("----Game NetSvr Failed-----")
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
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()

	// for k, v := range netConfig.G_SvrNetCfg {
	// 	fmt.Println(k, v)
	// }

	http.G_Before_Recv_Player = player.BeforeRecvNetMsg
	http.G_After_Recv_Player = player.AfterRecvNetMsg
}
