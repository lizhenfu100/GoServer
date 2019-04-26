package logic

import (
	"bytes"
	"common"
	"conf"
	"dbmgo"
	"encoding/json"
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"shared_svr/svr_sdk/msg"
)

func Http_order_info(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	orderId := q.Get("orderid")
	if p := msg.FindOrder(orderId); p != nil {
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	} else {
		w.Write(common.S2B(orderId + ": order not exists"))
	}
	gamelog.Info("Http_order_info: %v", r.Form)
}
func Http_order_success(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	orderId := q.Get("orderid")

	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
	} else if order := msg.FindOrder(orderId); order == nil {
		w.Write(common.S2B(orderId + ": order not exists"))
	} else if order.Status == 1 {
		w.Write(common.S2B(orderId + ": order already success"))
	} else {
		order.Status = 1
		order.Can_send = 1
		dbmgo.UpdateId(msg.KDBTable, order.Order_id, bson.M{"$set": bson.M{
			"status": 1, "can_send": 1}})
		w.Write(common.S2B("ok"))
	}
	gamelog.Info("Http_order_success: %v", r.Form)
}

// ------------------------------------------------------------
// 批量修改订单，配合内部工具使用
func Rpc_order_success(req, ack *common.NetPack) {
	var ackInfo bytes.Buffer
	for cnt, i := req.ReadUInt16(), uint16(0); i < cnt; i++ {
		orderId := req.ReadString()
		ackInfo.WriteString(orderId)

		if order := msg.FindOrder(orderId); order != nil {
			if order.Status == 1 {
				ackInfo.WriteString(": order already success\n")
			} else {
				order.Status = 1
				order.Can_send = 1
				dbmgo.UpdateId(msg.KDBTable, order.Order_id, bson.M{"$set": bson.M{
					"status": 1, "can_send": 1}})
				ackInfo.WriteString(": ok\n")
			}
		} else {
			ackInfo.WriteString(": order not exists\n")
		}
	}
	ack.WriteString(ackInfo.String())
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
