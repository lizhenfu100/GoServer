/***********************************************************************
* @ 单机游戏充值
* @ brief
    1、client先跟后台请求，生成充值订单
    2、client收到订单号，调第三方SDK下单，待回复后(SDK会同时通知client/server)，查询游戏后台能否发货
    3、先通知后台“发货确认”（Can_send = 0），成功后client再发货
    4、后续同一订单的发货查询随即失效
* @ author zhoumf
* @ date 2017-10-10
***********************************************************************/
package logic

import (
	"common/copy"
	"common/std/sign"
	"encoding/json"
	"fmt"
	"gamelog"
	"net/http"
	"shared_svr/svr_sdk/msg"
	"shared_svr/svr_sdk/platform"
)

func Http_pre_buy_request(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//gamelog.Debug("%v", r.Form)

	//反射解析订单信息
	var order msg.TOrderInfo
	copy.CopyForm(&order, r.Form)

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
	r.ParseForm()
	//gamelog.Debug("%v", r.Form)

	orderId := r.Form.Get("order_id")

	//! 创建回复
	ack := msg.Retcode_ack{Retcode: -1}
	var pResult interface{}
	pResult = &ack
	defer func() {
		b, _ := json.Marshal(pResult)
		w.Write(b)
		gamelog.Debug("ack: %v", pResult)
	}()

	if order := msg.FindOrder(orderId); order == nil {
		ack.Retcode = -2
	} else if r.Form.Get("sign") != sign.CalcSign("order_id="+order.Order_id) {
		ack.Retcode = -3
	} else if order.Status == 1 && order.Can_send == 1 {
		stOk := msg.Query_order_ack{}
		copy.CopySameField(&stOk.Order, order)
		pResult = &stOk
	}
}

//【先通知后台，再才发货，避免通知不成功重复发】
func Http_confirm_order(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//gamelog.Debug("%v", r.Form)

	orderId := r.Form.Get("order_id")

	//! 创建回复
	var ack msg.Retcode_ack
	ack.Retcode = -1
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	if order := msg.FindOrder(orderId); order == nil {
		ack.Retcode = -2
	} else if r.Form.Get("sign") != sign.CalcSign("order_id="+order.Order_id) {
		ack.Retcode = -3
	} else {
		ack.Retcode = 0
		msg.ConfirmOrder(order)
	}
}

func Http_query_order_unfinished(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	third := r.Form.Get("third_account")

	//! 创建回复
	var ack msg.Order_unfinished_ack
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()
	if third == "" {
		return
	}
	var order msg.UnfinishedOrder
	msg.OrderRange(func(k, v interface{}) bool {
		p := v.(*msg.TOrderInfo)
		if p.Third_account == third && p.Can_send == 1 {
			copy.CopySameField(&order, p)
			ack.Orders = append(ack.Orders, order)
		}
		return true
	})
}
