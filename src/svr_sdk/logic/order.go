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
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gamelog"
	"net/http"
	"strings"
	"svr_sdk/logic/kuaishou"
	"svr_sdk/msg"
)

// -------------------------------------
// 与平台约定的签名规则
const (
	k_pt_key = "yqqs(#(%$(%!$"
)

func calcSign(s string) string {
	key := fmt.Sprintf("%x", md5.Sum([]byte(k_pt_key)))
	sign := fmt.Sprintf("%s&%s", s, strings.ToLower(key))
	return fmt.Sprintf("%x", md5.Sum([]byte(sign)))
}

//客户端请求生成订单
func Http_pre_buy_request(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//反射解析订单信息
	var order msg.TOrderInfo
	msg.Unmarshal(&order, r.Form)

	//! 创建回复
	ack := NewPreBuyAck(order.Pf_id)
	ack.SetRetcode(-1)
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	//验证签名
	s := fmt.Sprintf("pf_id=%s&pk_id=%s&pay_id=%d&item_id=%d&item_count=%d&total_price=%d", order.Pf_id, order.Pk_id, order.Pay_id, order.Item_id, order.Item_count, order.Total_price)
	if r.Form.Get("sign") != calcSign(s) {
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

	//需后台下单的渠道
	switch order.Pf_id {
	case "kuaishou":
		if !kuaishou.OrderToThirdparty(&order, ack) {
			return
		}
	}
	ack.SetOrderId(order.Order_id)
	ack.SetRetcode(0)
}
func NewPreBuyAck(pf_id string) msg.Pre_buy_ack {
	switch pf_id {
	case "kuaishou":
		return &kuaishou.Pre_buy_ack{}
	}
	return &msg.Pre_buy{}
}

//客户端查询订单，是否购买成功、是否发货过
func Http_query_order(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//! 创建回复
	var ack msg.Query_order_ack
	ack.Retcode = -1
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	print(r.Form["order_id"][0])
	order := msg.FindOrder(r.Form.Get("order_id"))
	if order == nil {
		gamelog.Debug("none order_id: %s", r.Form.Get("order_id"))
		return
	}
	if r.Form.Get("sign") != calcSign("order_id="+order.Order_id) { //验证签名
		gamelog.Error("Rpc_sdk_query_order: sign failed")
		return
	}

	if order.Status == 1 && order.Can_send == 1 {
		ack.Retcode = 0
		//回复订单信息
		msg.CopySameField(&ack.Order, order)
	}
}

//客户端发货成功，通告后台，避免重复发货
func Http_confirm_order(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//! 创建回复
	var ack msg.Retcode_ack
	ack.Retcode = -1
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	if order := msg.FindOrder(r.Form.Get("order_id")); order != nil {
		if r.Form.Get("sign") != calcSign("order_id="+order.Order_id) { //验证签名
			gamelog.Error("Rpc_sdk_confirm_order: sign failed")
			return
		}
		ack.Retcode = 0
		msg.ConfirmOrder(order)
	}
}
