/***********************************************************************
* @ 游戏服msg的处理
* @ brief
    1、gamesvr先通知SDK进程，建立新充值订单

    2、第三方充值信息到达后，验证是否为有效订单

* @ author zhoumf
* @ date 2016-8-18
***********************************************************************/
package logic

import (
	"common"
	"encoding/json"
	"fmt"
	"gamelog"
	"msg/sdk_msg"
	"net/http"
)

func Rpc_sdk_echo(req, ack *common.NetPack) {
}
func Rpc_sdk_notify_recharge_order(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	//! 解析消息
	var req sdk_msg.SDKMsg_create_recharge_order_Req
	err := json.Unmarshal(buffer, &req)
	if err != nil {
		gamelog.Error("Rpc_Notify_Recharge_Order unmarshal fail. Error: %s", err.Error())
		return
	}

	fmt.Println(req)

	//! 创建回复
	var response sdk_msg.SDKMsg_create_recharge_order_Ack
	response.RetCode = -1
	defer func() {
		b, _ := json.Marshal(&response)
		w.Write(b)
	}()

	//TODO：生成订单，等待第三方的充值信息
	pOrder := new(TRechargeOrder)
	pOrder.OrderID = req.OrderID
	pOrder.GamesvrID = req.GamesvrID
	pOrder.PlayerID = req.PlayerID
	pOrder.AccountID = req.PlayerID
	pOrder.Channel = req.Channel
	pOrder.PlatformEnum = req.PlatformEnum
	pOrder.chargeCsvID = req.ChargeCsvID
	CacheRechargeOrder(pOrder)

	response.RetCode = 0
}
