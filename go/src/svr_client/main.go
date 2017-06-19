package main

import (
	"common"
	"conf"
	"dbmgo"
	"fmt"
	"gamelog"
	"http"
	"netConfig"
	"time"
	//"msg/sdk_msg"

	//"svr_client/api"
	//"svr_game/logic/player"
)

func main() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()

	//初始化日志系统
	gamelog.InitLogger("client")
	gamelog.SetLevel(0)

	dbmgo.Init(conf.GameDbAddr, conf.GameDbName)

	// for k, v := range netConfig.G_SvrNetCfg {
	// 	fmt.Println(k, v)
	// }
	netConfig.CreateNetSvr("client", 0)

	test()
	time.Sleep(100 * time.Second)
}

func test() {
	gameAddr := netConfig.GetHttpAddr("game", -1)
	centerAddr := netConfig.GetHttpAddr("center", -1)
	fmt.Println("---", gameAddr)
	fmt.Println("---", centerAddr)
	gameRpc := http.NewClientRpc(gameAddr, 0)
	centerRpc := http.NewClientRpc(centerAddr, 0)

	//向游戏服请求充值
	// var msg1 sdk_msg.Msg_create_recharge_order_Req
	// msg1.SessionKey = "233xx"
	// msg1.OrderID = "abcdefg233"
	// msg1.Channel = "360"
	// msg1.PlatformEnum = 0
	// msg1.ChargeCsvID = 2
	// buf1, _ := json.Marshal(&msg1)
	// http.PostReq(gameAddr+"create_recharge_order", buf1)

	// //模拟第三方的充值到账
	// sdkAddr := netConfig.GetHttpAddr("sdk", -1)
	// fmt.Println("---", sdkAddr)
	// var msg2 sdk_msg.SDKMsg_recharge_result
	// msg2.OrderID = "abcdefg233"
	// msg2.ThirdOrderID = "zzzzzzzzz"
	// msg2.Channel = "360"
	// msg2.RMB = 233
	// buf2, _ := json.Marshal(&msg2)
	// http.PostReq(sdkAddr+"sdk_recharge_success", buf2)

	time.Sleep(2 * time.Second)

	//向center取游戏服务器列表
	accountName := "zhoumf"
	password := "123"
	accountId := uint32(0)
	gamesvrIp := ""
	gamesvrPort := uint16(0)
	logintoken := uint32(0)

	centerRpc.CallRpc("rpc_center_reg_account", func(buf *common.NetPack) {
		buf.WriteString(accountName)
		buf.WriteString(password)
	}, func(backBuf *common.NetPack) {
		errCode1 := backBuf.ReadInt8()
		fmt.Println("errCode1:", errCode1)
	})

	centerRpc.CallRpc("rpc_center_login_gamesvr", func(buf *common.NetPack) {
		buf.WriteString(accountName)
		buf.WriteString(password)
		buf.WriteInt(1) //gamesvrId
	}, func(backBuf *common.NetPack) {
		errCode2 := backBuf.ReadInt8()
		if errCode2 > 0 {
			accountId = backBuf.ReadUInt32()
			gamesvrIp = backBuf.ReadString()
			gamesvrPort = backBuf.ReadUInt16()
			logintoken = backBuf.ReadUInt32()
		} else {
			fmt.Println("errCode2:", errCode2)
		}
	})

	//登录
	gameRpc.CallRpc("rpc_game_login", func(buf *common.NetPack) {
		buf.WriteUInt32(accountId)
		buf.WriteUInt32(logintoken)
	}, func(backBuf *common.NetPack) {
		errCode3 := backBuf.ReadInt8()
		if errCode3 > 0 {
			gameRpc.PlayerId = backBuf.ReadUInt32()
		} else if errCode3 == -2 {
			//创建新角色
			gameRpc.CallRpc("rpc_game_player_create", func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteString("zhoumf")
			}, func(backBuf *common.NetPack) {
				gameRpc.PlayerId = backBuf.ReadUInt32()
			})
		} else {
			fmt.Println("errCode3:", errCode3)
		}
	})

	//登出
	// gameRpc.CallRpc("rpc_game_logout", func(buf *common.NetPack) {
	// }, func(backBuf *common.NetPack) {
	// })

	// time.Sleep(2 * time.Second)
	// //直接发给战斗服
	// msg := common.NewNetPackCap(32)
	// msg.SetRpc("rpc_echo")
	// msg.WriteString("--- zhoumf 233 --- ")
	// api.SendToBattle(1, msg)
	// api.SendToBattle(2, msg)

	time.Sleep(2 * time.Hour)
}
