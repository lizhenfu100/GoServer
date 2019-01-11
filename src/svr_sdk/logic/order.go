/***********************************************************************
* @ 单机游戏充值
* @ brief
    1、client先向后台请求，生成充值订单
    2、client得到订单号后，通知第三方SDK，待其回复后(SDK回同时通知client/server)，查询游戏后台能否发货
    3、发货后，通告后台“发货确认”，后台将发货标记 Can_send 置空
    4、后续同一订单的发货查询随即失效

* @ author zhoumf
* @ date 2017-10-10
***********************************************************************/
package logic

import (
	"bytes"
	"common"
	"common/sign"
	"dbmgo"
	"encoding/json"
	"fmt"
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"svr_sdk/msg"
	"svr_sdk/platform"
)

func Http_pre_buy_request(w http.ResponseWriter, r *http.Request) {
	//gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//反射解析订单信息
	var order msg.TOrderInfo
	common.Unmarshal(&order, r.Form)

	//! 创建回复
	ack := platform.NewPreBuyAck(order.Pf_id)
	ack.SetRetcode(-1)
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	//验证签名
	s := fmt.Sprintf("pf_id=%s&pk_id=%s&pay_id=%d&item_id=%d&item_count=%d&total_price=%d",
		order.Pf_id, order.Pk_id, order.Pay_id, order.Item_id, order.Item_count, order.Total_price)
	if r.Form.Get("sign") != sign.CalcSign(s) {
		ack.SetRetcode(-2)
		gamelog.Error("pre_buy_request: sign failed")
		return
	}
	//生成订单
	if !msg.CreateOrderInDB(&order) {
		ack.SetRetcode(-3)
		gamelog.Error("pre_buy_request: create order failed")
		return
	}

	//如需后台下单的，在各自HandleOrder()中处理
	if ack.HandleOrder(&order) {
		ack.SetOrderId(order.Order_id)
		ack.SetRetcode(0)
	}
}

//客户端查询订单，是否购买成功、是否发货过
func Http_query_order(w http.ResponseWriter, r *http.Request) {
	//gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//! 创建回复
	var ack msg.Query_order_ack
	ack.Retcode = -1
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)

		if ack.Retcode != 0 {
			gamelog.Debug("ack: %v, req: %s", ack, r.URL.String())
		}
	}()

	order := msg.FindOrder(r.Form.Get("order_id"))
	if order == nil {
		ack.Retcode = -2
		return
	}
	if r.Form.Get("sign") != sign.CalcSign("order_id="+order.Order_id) { //验证签名
		ack.Retcode = -3
		return
	}

	if order.Status == 1 && order.Can_send == 1 {
		ack.Retcode = 0
		//回复订单信息
		common.CopySameField(&ack.Order, order)
	}
}

//客户端发货成功，通告后台，避免重复发货
func Http_confirm_order(w http.ResponseWriter, r *http.Request) {
	//gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//! 创建回复
	var ack msg.Retcode_ack
	ack.Retcode = -1
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)

		if ack.Retcode != 0 {
			gamelog.Debug("ack: %v, req: %s", ack, r.URL.String())
		}
	}()

	if order := msg.FindOrder(r.Form.Get("order_id")); order != nil {
		if r.Form.Get("sign") != sign.CalcSign("order_id="+order.Order_id) { //验证签名
			ack.Retcode = -3
			return
		}
		ack.Retcode = 0
		msg.ConfirmOrder(order)
	} else {
		ack.Retcode = -2
	}
}

// --------------------------------------------------------------------------
// 运维用，修改订单
func Rpc_order_success(req, ack *common.NetPack) {
	var errInfo bytes.Buffer
	cnt := req.ReadUInt16()
	for i := uint16(0); i < cnt; i++ {
		orderId := req.ReadString()

		if order := msg.FindOrder(orderId); order != nil {
			if order.Status == 1 {
				errInfo.WriteString(orderId)
				errInfo.WriteString(": order already success\n")
			} else {
				order.Status = 1
				order.Can_send = 1
				dbmgo.UpdateId("Order", order.Order_id, bson.M{"$set": bson.M{"status": 1, "can_send": 1}})
			}
		} else {
			errInfo.WriteString(orderId)
			errInfo.WriteString(": order not exists\n")
		}
	}

	ack.WriteString(errInfo.String())
}
func Rpc_order_info(req, ack *common.NetPack) {
	cnt := req.ReadUInt16()
	ack.WriteUInt16(cnt)
	for i := uint16(0); i < cnt; i++ {
		orderId := req.ReadString()

		if order := msg.FindOrder(orderId); order != nil {
			ack.WriteInt8(1)
			ack.WriteString(order.Third_order_id)
			ack.WriteString(order.Third_account)
			ack.WriteString(order.Item_name)
			ack.WriteInt(order.Item_count)
			ack.WriteInt(order.Total_price)
			ack.WriteInt64(order.Time)
			ack.WriteInt(order.Status)
			ack.WriteInt(order.Can_send)
		} else {
			ack.WriteInt8(-1)
		}
	}
}
