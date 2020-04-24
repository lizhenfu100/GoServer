package logic

import (
	"common"
	"conf"
	"dbmgo"
	"encoding/json"
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"shared_svr/svr_sdk/msg"
	"strings"
)

func Http_order_info(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	orderId := r.Form.Get("orderid")
	if p := msg.FindOrder(orderId); p != nil {
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	} else {
		w.Write(common.S2B(orderId + ": order not exists"))
	}
}
func Http_order_success(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm(); r.Form.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	for _, v := range strings.Split(r.Form.Get("orderid"), ",") {
		if order := msg.FindOrder(v); order == nil {
			w.Write(common.S2B(v + ": order not exists"))
		} else if order.Status == 1 {
			w.Write(common.S2B(v + ": order already success"))
		} else {
			order.Status = 1
			order.Can_send = 1
			dbmgo.UpdateIdSync(msg.KDBTable, order.Order_id, bson.M{"$set": bson.M{
				"status": 1, "can_send": 1}})
			w.Write([]byte("ok"))
		}
		w.Write([]byte("\n"))
	}
	gamelog.Info("Http_order_success: %v", r.Form)
}
func Http_order_set_force(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm(); r.Form.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	for _, v := range strings.Split(r.Form.Get("orderid"), ",") {
		if order := msg.FindOrder(v); order == nil {
			w.Write(common.S2B(v + ": order not exists"))
		} else {
			order.Status = 1
			order.Can_send = 1
			dbmgo.UpdateIdSync(msg.KDBTable, order.Order_id, bson.M{"$set": bson.M{
				"status": 1, "can_send": 1}})
			w.Write([]byte("ok"))
		}
		w.Write([]byte("\n"))
	}
	gamelog.Info("Http_order_set_force: %v", r.Form)
}
