package huluxia

import (
	"dbmgo"
	"gamelog"
	"net/http"

	"gopkg.in/mgo.v2/bson"
	"svr_sdk/logic"
	"svr_sdk/msg"
)

//第三方通告订单成功
func Http_huluxia_order_success(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//! 创建回复
	ack := "success"
	defer func() {
		w.Write([]byte(ack))
		gamelog.Debug("huluxia_order_success ack: %s", ack)
	}()

	// 解析参数
	var req RechargeSuccess_req
	msg.Unmarshal(&req, r.Form)

	if req.Sign != req.CalcSign() {
		ack = "fail"
		gamelog.Debug("%v \n sign:%s", req, req.CalcSign())
		return
	}
	order := logic.FindOrder(req.Out_order_no)
	if order == nil {
		ack = "fail"
		gamelog.Debug("order_id(%s) not exists", req.Out_order_no)
		return
	}
	if order.Third_account != req.Uid {
		ack = "fail"
		gamelog.Debug("third_account:%s, uid:%s", order.Third_account, req.Uid)
		return
	}
	if order.Status == 1 {
		ack = "fail"
		gamelog.Debug("order repeat")
		return
	}

	order.Status = 1
	order.Can_send = 1
	order.Third_order_id = req.Order_no
	dbmgo.UpdateToDB("Order", bson.M{"_id": order.Order_id}, bson.M{"$set": bson.M{"third_order_id": order.Third_order_id, "status": 1, "can_send": 1}})
}
