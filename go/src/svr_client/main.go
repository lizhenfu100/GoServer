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
	"svr_game/logic/player"
)

func main() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	//初始化日志系统
	gamelog.InitLogger("client")
	gamelog.SetLevel(0)

	dbmgo.Init(conf.GameDbAddr, conf.GameDbName)

	common.LoadAllCsv()
	// for k, v := range netConfig.G_SvrNetCfg {
	// 	fmt.Println(k, v)
	// }
	netConfig.CreateNetSvr("client", 0)

	test()
	time.Sleep(100 * time.Second)
}

func test() {
	//向游戏服请求充值
	gameAddr := netConfig.GetHttpAddr("game", -1)
	centerAddr := netConfig.GetHttpAddr("center", -1)
	fmt.Println("---", gameAddr)
	fmt.Println("---", centerAddr)
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
	sendBuf := common.NewByteBufferCap(32)
	// sendBuf.WriteByte(1)
	// http.PostReq(gameAddr+"rpc_test_mongodb", sendBuf.DataPtr)

	//向center取游戏服务器列表
	accountName := "zhoumf"
	password := "123"
	accountId := uint32(0)
	accountBuf := common.NewByteBufferCap(32)
	accountBuf.WriteString(accountName)
	accountBuf.WriteString(password)
	{
		b := http.PostReq(centerAddr+"rpc_reg_account", accountBuf.DataPtr)
		buf := common.NewByteBuffer(b)
		errCode1 := buf.ReadInt8()
		fmt.Println("errCode1:", errCode1)
	}
	{
		b := http.PostReq(centerAddr+"rpc_get_gamesvr_lst", accountBuf.DataPtr)
		buf := common.NewByteBuffer(b)
		errCode2 := buf.ReadInt8()
		if errCode2 > 0 {
			accountId = buf.ReadUInt32()
			svrId := buf.ReadUInt32()
			size := buf.ReadByte()
			for i := byte(0); i < size; i++ {
				module := buf.ReadString()
				id := buf.ReadUInt32()
				ip := buf.ReadString()
				port := buf.ReadUInt16()
				fmt.Println("GameSvr:", module, id, ip, port)
			}
			fmt.Println("Account Info:", accountId, svrId)
		} else {
			fmt.Println("errCode2:", errCode2)
		}
	}

	//向游戏服发战斗数据，后台game转到battle
	// buf.ClearBody()
	// buf.WriteString("client-game-battle")
	// http.PostReq(gameAddr+"battle_echo", buf.DataPtr)

	//创建
	// {
	// 	sendBuf.ClearBody()
	// 	sendBuf.WriteUInt32(accountId)
	// 	sendBuf.WriteString("zhoumf")
	// 	b := http.PostReq(gameAddr+"rpc_player_create", sendBuf.DataPtr)
	// 	buf := common.NewByteBuffer(b)
	// 	playerId := buf.ReadUInt32()

	// 	//写邮件
	// 	player := player.FindWithDB_PlayerId(playerId)
	// 	fmt.Println("create player id:", playerId, "\n", player)
	// }
	//登录
	{
		sendBuf.ClearBody()
		sendBuf.WriteUInt32(accountId)
		b := http.PostReq(gameAddr+"rpc_player_login", sendBuf.DataPtr)
		buf := common.NewByteBuffer(b)
		fmt.Println("rpc_player_login:", b)
		// 先读逻辑回包

		// http svr send data
		bit := buf.ReadUInt32()
		if common.GetBit32(bit, player.Bit_Mail_Lst) {
			var mail player.TMail
			cnt := buf.ReadUInt32()
			for i := uint32(0); i < cnt; i++ {
				mail.BufToMail(buf)
				fmt.Println(mail)
			}
		}
	}

	// time.Sleep(2 * time.Second)
	// //直接发给战斗服
	// msg := common.NewNetPackCap(32)
	// msg.SetRpc("rpc_echo")
	// msg.WriteString("--- zhoumf 233 --- ")
	// api.SendToBattle(1, msg)
	// api.SendToBattle(2, msg)
}
