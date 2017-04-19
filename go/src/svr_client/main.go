package main

import (
	"common"
	//"encoding/json"
	"fmt"
	"gamelog"
	"http"
	"netConfig"
	//"svr_client/api"
	"time"
	//"msg/sdk_msg"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("client")
	gamelog.SetLevel(0)

	RegClientCsv()
	// for k, v := range netConfig.G_SvrNetCfg {
	// 	fmt.Println(k, v)
	// }

	netConfig.CreateNetSvr("client", 0)

	test()
	time.Sleep(100 * time.Second)
}

func RegClientCsv() {
	var config = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	//! register
	for k, v := range config {
		common.G_CsvParserMap[k] = v
	}
	common.LoadAllCsv()
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
	// http.PostReq(gameAddr+"/create_recharge_order", buf1)

	// //模拟第三方的充值到账
	// sdkAddr := netConfig.GetHttpAddr("sdk", -1)
	// fmt.Println("---", sdkAddr)
	// var msg2 sdk_msg.SDKMsg_recharge_result
	// msg2.OrderID = "abcdefg233"
	// msg2.ThirdOrderID = "zzzzzzzzz"
	// msg2.Channel = "360"
	// msg2.RMB = 233
	// buf2, _ := json.Marshal(&msg2)
	// http.PostReq(sdkAddr+"/sdk_recharge_success", buf2)

	time.Sleep(2 * time.Second)
	//向游戏服发战斗数据，后台game转到battle
	buf := common.NewNetPackCap(32)
	buf.SetRpc("rpc_echo")
	buf.WriteString("client-game-battle")
	b, _ := http.PostReq(gameAddr+"/battle_echo", buf.DataPtr)
	fmt.Println("---", string(b))

	buf.ClearBody()
	buf.WriteByte(4)
	http.PostReq(gameAddr+"/rpc_test_mongodb", buf.DataPtr)

	//向center取游戏服务器列表
	{
		b, err := http.PostReq(centerAddr+"/rpc_get_gamesvr_lst", []byte{})
		if err != nil {
			fmt.Println("Error:", err)
		}
		buf := common.NewNetPack(b)
		size := buf.ReadByte()
		for i := byte(0); i < size; i++ {
			module := buf.ReadString()
			id := buf.ReadInt()
			ip := buf.ReadString()
			port := buf.ReadInt()
			fmt.Println("GameSvr:", module, id, ip, port)
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
