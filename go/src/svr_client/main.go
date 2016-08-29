package main

import (
	"encoding/json"
	"fmt"
	"gamelog"
	"http"
	"netConfig"

	"msg/sdk_msg"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("client", true)
	gamelog.SetLevel(0)

	netConfig.CreateNetSvr("client")

	gameAddr := netConfig.GetHttpAddr("game", 0)
	fmt.Println("---", gameAddr)
	var msg1 sdk_msg.Msg_create_recharge_order_Req
	msg1.SessionKey = "233xx"
	msg1.OrderID = "abcdefg233"
	msg1.Channel = "360"
	msg1.PlatformEnum = 0
	msg1.ChargeCsvID = 2
	buf1, _ := json.Marshal(&msg1)
	http.PostReq(gameAddr+"/create_recharge_order", buf1)

	sdkAddr := netConfig.GetHttpAddr("sdk", 0)
	fmt.Println("---", sdkAddr)
	var msg2 sdk_msg.SDKMsg_recharge_result
	msg2.OrderID = "abcdefg233"
	msg2.ThirdOrderID = "zzzzzzzzz"
	msg2.Channel = "360"
	msg2.RMB = 233
	buf2, _ := json.Marshal(&msg2)
	http.PostReq(sdkAddr+"/sdk_recharge_success", buf2)

	http.PostReq(gameAddr+"/battle_echo", []byte("zhoumf 233!!!"))
}
