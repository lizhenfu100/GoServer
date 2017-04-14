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
	// fmt.Println("---", gameAddr)
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
	buf := common.NewNetPack(32)
	buf.SetOpCode(0)
	buf.WriteString("client-game-battle")
	b, _ := http.PostReq(gameAddr+"/battle_echo", buf.DataPtr)
	fmt.Println("---", b)

	// time.Sleep(2 * time.Second)
	// //直接发给战斗服
	// msg := common.NewNetPack(32)
	// msg.SetOpCode(1)
	// msg.WriteString("--- zhoumf 233 --- ")
	// api.SendToBattle(1, msg)
	// api.SendToBattle(2, msg)
}
