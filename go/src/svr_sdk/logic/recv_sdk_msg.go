/***********************************************************************
* @ 第三方msg的业务处理
* @ brief
    1、已通过验证，消息为JSON格式，解析后转发给对应gamesvr

    2、如是订单信息，还要验证订单号，是否已在SDK注册(gamesvr会先通知SDK建立新订单)
        否则，第三方和我们内部的数据可能不匹配

    3、订单信息解析后立即写库，回复第三方ok，不管gamesvr是否成功

    4、如gamesvr充值失败，走客服补单流程…………到手的钱不能飞了~( ▔___▔)y

* @ author zhoumf
* @ date 2016-8-18
***********************************************************************/
package logic

import (
	"encoding/json"
	"gamelog"
	"msg/sdk_msg"
	"net/http"
	"svr_sdk/api"
)

func HandSdk_RechargeSuccess(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	//! 解析消息
	var req sdk_msg.SDKMsg_recharge_result
	err := json.Unmarshal(buffer, &req)
	if err != nil {
		gamelog.Error("HandSdk_RechargeSuccess unmarshal fail. Error: %s", err.Error())
		return
	}

	//! 创建回复
	response := "fail"
	defer func() {
		w.Write([]byte(response)) //Notice：用defer安全些，但得等RelayToGamesvr返回，会慢一点
	}()

	//TODO：验证token，解析JSON数据
	jsonContent := ""
	if CheckToken(req.Channel) == false {
		//TODO：回复第三方错误信息
		response = "error info"
		return
	}

	//TODO：验证订单，入库，回第三方ok
	if pOrder := DB_Save_RechargeOrder(req.OrderID, req.ThirdOrderID, jsonContent, req.RMB); pOrder != nil {
		response = "ok"
		//TODO：通知gamesvr，充值成功
		var gamesvrReq sdk_msg.Msg_recharge_success
		gamesvrReq.PlayerID = req.PlayerID
		gamesvrReq.ChargeCsvID = pOrder.chargeCsvID
		gamesvrReq.RMB = req.RMB
		api.RelayToGamesvr(pOrder.GamesvrID, "sdk_recharge_success", &gamesvrReq)
	}
}
