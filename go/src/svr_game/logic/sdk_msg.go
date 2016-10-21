/***********************************************************************
* @ 游戏服通知SDK进程
* @ brief
    1、gamesvr先通知SDK进程，建立新充值订单

    2、第三方充值信息到达后，验证是否为有效订单

* @ author zhoumf
* @ date 2016-8-18
***********************************************************************/
package logic

import (
	"encoding/json"
	"fmt"
	"gamelog"
	myhttp "http"
	"msg/sdk_msg"
	"net/http"
	"netConfig"
)

var (
	g_cache_sdk_addr string
)

// strKey = "/create_recharge_order"
func PostSdkReq(strKey string, pMsg interface{}) ([]byte, error) {
	if g_cache_sdk_addr == "" {
		g_cache_sdk_addr = netConfig.GetHttpAddr("sdk", 0)
	}

	buf, _ := json.Marshal(pMsg)
	url := g_cache_sdk_addr + strKey

	return myhttp.PostReq(url, buf)
}

//! 消息处理函数
//
func Handle_Create_Recharge_Order(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	//! 解析消息
	var req sdk_msg.Msg_create_recharge_order_Req
	err := json.Unmarshal(buffer, &req)
	if err != nil {
		gamelog.Error("Handle_Create_Recharge_Order unmarshal fail. Error: %s", err.Error())
		return
	}

	fmt.Println(req)

	//! 创建回复
	var response sdk_msg.Msg_create_recharge_order_Ack
	response.RetCode = -1
	defer func() {
		b, _ := json.Marshal(&response)
		w.Write(b)
	}()

	//! 常规检查
	// var player *TPlayer = nil
	// player, response.RetCode = GetPlayerAndCheck(req.PlayerID, req.SessionKey, r.URL.String())
	// if player == nil {
	// 	return
	// }

	// 转发给SDK进程
	var sdkReq sdk_msg.SDKMsg_create_recharge_order_Req
	var sdkAck sdk_msg.SDKMsg_create_recharge_order_Ack
	sdkReq.GamesvrID = netConfig.G_Local_SvrID
	sdkReq.PlayerID = req.PlayerID
	sdkReq.OrderID = req.OrderID
	sdkReq.Channel = req.Channel
	sdkReq.PlatformEnum = req.PlatformEnum
	sdkReq.ChargeCsvID = req.ChargeCsvID
	backBuf, err := PostSdkReq("/create_recharge_order", &sdkReq)
	if err == nil {
		json.Unmarshal(backBuf, &sdkAck)
		//TODO：将SDKMsg_create_recharge_order_Ack中的数据，写入response
	} else {
		fmt.Println(err.Error())
	}

	// 回复client，client会将订单信息发给第三方
	response.RetCode = 0
}
func Handle_Recharge_Success(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	//! 解析消息
	var req sdk_msg.Msg_recharge_success
	err := json.Unmarshal(buffer, &req)
	if err != nil {
		gamelog.Error("Handle_Recharge_Success unmarshal fail. Error: %s", err.Error())
		return
	}

	//! 创建回复
	defer func() {
		w.Write([]byte("ok"))
	}()

	// 充值到账，增加钻石数量
	// var player *TPlayer = GetPlayerByID(req.PlayerID)
	// if player == nil {
	// 	gamelog.Error("Handle_Recharge_Success GetPlayerByID nil! Invalid Player ID:%d, ChargeCsvID:%d, RMB:%d", req.PlayerID, req.ChargeCsvID, req.RMB)
	// 	return
	// }
	// player.HandChargeRenMinBi(req.RMB, req.ChargeCsvID)
}
