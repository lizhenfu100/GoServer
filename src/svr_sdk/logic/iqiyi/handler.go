package iqiyi

import (
	"dbmgo"
	"encoding/json"
	"fmt"
	"gamelog"
	"io/ioutil"
	"net/http"
	"net/url"

	"gopkg.in/mgo.v2/bson"
	"svr_sdk/logic"
	"svr_sdk/msg"
)

func VoucherOrders(timeBegin, timeEnd string) *VoucherOrder_ack { //2006-01-02 15:04:05
	req := VoucherOrder_req{
		Game_id:    "6553",
		Start_time: timeBegin,
		End_time:   timeEnd,
	}
	u, _ := url.Parse("http://pay.game.iqiyi.com/interface/orderinfo/voucherOrders")
	q := u.Query()
	q.Set("game_id", req.Game_id)
	q.Set("start_time", req.Start_time)
	q.Set("end_time", req.End_time)
	//q.Set("order_type", req.Order_type)
	q.Set("sign", req.calcSign())
	u.RawQuery = q.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		return nil
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil
	}

	var ack VoucherOrder_ack
	json.Unmarshal(result, &ack)
	if ack.Code != 0 {
		return nil
	}
	return &ack
}

//第三方通告订单成功
func Http_iqiyi_order_success(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//! 创建回复
	var ack RechargeSuccess_ack
	ack.Result = -6
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("iqiyi_order_success ack: %v", ack)
	}()

	// 解析参数
	var req RechargeSuccess_req
	msg.Unmarshal(&req, r.Form)

	if req.Sign != req.calcSign() {
		ack.Result = -1
		ack.Message = "sing:" + req.calcSign()
		return
	}
	//client把我们生成的order_id传给第三方的
	order := logic.FindOrder(req.UserData)
	if order == nil {
		ack.Result = -2
		ack.Message = fmt.Sprintf("order_id(%s) not exists", req.UserData)
		return
	}
	if order.Third_account != req.User_id {
		ack.Result = -3
		return
	}
	if order.Status == 1 {
		ack.Result = -4
		ack.Message = "order repeat"
		return
	}

	order.Status = 1
	order.Can_send = 1
	order.Third_order_id = req.Order_id
	dbmgo.UpdateToDB("Order", bson.M{"_id": order.Order_id}, bson.M{"$set": bson.M{"third_order_id": order.Third_order_id, "status": 1, "can_send": 1}})
	ack.Result = 0
}
