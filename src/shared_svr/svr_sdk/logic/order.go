/***********************************************************************
* @ 单机游戏充值
* @ brief
    1、client先跟后台请求，生成充值订单
    2、client收到订单号，调第三方SDK下单，待回复后(SDK会同时通知client/server)，查询游戏后台能否发货
    3、先通知后台“发货确认”（Can_send = 0），成功后client再发货
    4、后续同一订单的发货查询随即失效

* @ 内网联调
	1、路由器-NAT-虚拟服务器，映射本机端口
	2、查看路由器界面显示的ip，不能是web返回的本地ip
	3、将路由器ip、本机端口，发给第三方

* @ author zhoumf
* @ date 2017-10-10
***********************************************************************/
package logic

import (
	"common"
	"common/assert"
	"common/copy"
	"common/safe"
	"common/std/sign"
	"dbmgo"
	"encoding/json"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"netConfig"
	"shared_svr/svr_gm/conf"
	"shared_svr/svr_sdk/msg"
	"shared_svr/svr_sdk/platform"
	"sync"
)

func Http_pre_buy_request(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gamelog.Debug("pre_buy: %v", r.Form)

	//反射解析订单信息
	var order msg.TOrderInfo
	copy.CopyForm(&order, r.Form)

	//! 创建回复
	ack := platform.NewPreBuyAck(order.Pay_id)
	ack.SetRetcode(-1, "")
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	//验证签名
	s := fmt.Sprintf("pf_id=%s&pk_id=%s&pay_id=%d&item_id=%d&item_count=%d&total_price=%d",
		order.Pf_id, order.Pk_id, order.Pay_id, order.Item_id, order.Item_count, order.Total_price)
	if !CheckSign(r.Form.Get("sign"), s, order.Game_id) {
		ack.SetRetcode(-2, "sign failed")
		gamelog.Error("pre_buy_request: sign failed")
		return
	}
	//是否被封
	if !CheckMac(order.Mac) {
		ack.SetRetcode(-4, "Mac is banned")
		return
	}
	//生成订单
	if e := msg.CreateOrderInDB(&order); e != nil {
		ack.SetRetcode(-3, "create order failed")
		gamelog.Error(e.Error())
		return
	}
	//如需后台下单的，在各自HandleOrder()中处理
	if ack.HandleOrder(&order) {
		ack.SetOrderId(order.Order_id)
		ack.SetRetcode(0, "")
		AddMacOrder(order.Mac, order.Order_id) //统计Mac下的无效订单
	}
}

//客户端查询订单，是否购买成功、是否发货过
func Http_query_order(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gamelog.Debug("query_order: %v", r.Form)

	orderId := r.Form.Get("order_id")

	//! 创建回复
	ack := msg.Retcode_ack{-1, "unpaid"}
	var pResult interface{}
	pResult = &ack
	defer func() {
		b, _ := json.Marshal(pResult)
		w.Write(b)
		gamelog.Debug("ack: %v", pResult)
	}()

	if order := msg.FindOrder(orderId); order == nil {
		ack.Retcode = -2
		ack.Msg = "order none"
	} else if !CheckSign(r.Form.Get("sign"), "order_id="+orderId, order.Game_id) {
		ack.Retcode = -3
		ack.Msg = "sign failed"
	} else if order.Status == 1 && order.Can_send == 1 {
		stOk := msg.Query_order_ack{}
		copy.CopySameField(&stOk.Order, order)
		pResult = &stOk
	}
}

//先通知后台，再才发货，避免通知不成功重复发
func Http_confirm_order(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gamelog.Debug("confirm: %v", r.Form)

	orderId := r.Form.Get("order_id")

	//! 创建回复
	ack := msg.Retcode_ack{-1, ""}
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	if order := msg.FindOrder(orderId); order == nil {
		ack.Retcode = -2
		ack.Msg = "order none"
	} else if !CheckSign(r.Form.Get("sign"), "order_id="+orderId, order.Game_id) {
		ack.Retcode = -3
		ack.Msg = "sign failed"
	} else {
		ack.Retcode = 0
		platform.ConfirmOrder(order)
		DelMacOrder(order.Mac, order.Order_id) //统计Mac下的无效订单
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
	}()
	if third == "" {
		return
	}
	var list []msg.TOrderInfo
	dbmgo.FindAll(msg.KDBTable, bson.M{
		"third_account": third,
		"can_send":      1,
	}, &list)
	var order msg.UnfinishedOrder
	for i := 0; i < len(list); i++ {
		copy.CopySameField(&order, &list[i])
		ack.Orders = append(ack.Orders, order)
	}
}

// ------------------------------------------------------------
func CheckSign(_sign, s, gameid string) bool {
	if gameid == "SoulKnight" {
		const k_sign_key = "xdc*ef&xzz"
		return _sign == sign.CalcSign(s) || _sign == sign.GetSign(s, k_sign_key)
	}
	return _sign == sign.CalcSign(s)
}

// ------------------------------------------------------------
// 封Mac
var g_mac_order sync.Map //<mac, orderIds> //设备下的订单

func CheckMac(mac string) bool {
	ret := byte(1)
	if p, ok := netConfig.GetRpcRand("gm"); ok {
		isSyncCall := true
		p.CallRpc(enum.Rpc_gm_forbid_check, func(buf *common.NetPack) {
			isSyncCall = false
			buf.WriteString(conf.Order_Mac)
			buf.WriteString(mac)
		}, func(recvbuf *common.NetPack) {
			isSyncCall = true
			ret = recvbuf.ReadByte()
		})
		assert.True(isSyncCall)
	}
	return ret > 0
}
func AddMacOrder(mac, id string) {
	if ids, ok := g_mac_order.Load(mac); ok {
		ids.(*safe.Strings).Add(id)
		if n := ids.(*safe.Strings).Size(); n > 5 {
			gamelog.Info("invalid order cnt: %s(%d)", mac, n)
			if p, ok := netConfig.GetRpcRand("gm"); ok {
				p.CallRpc(enum.Rpc_gm_forbid_add, func(buf *common.NetPack) {
					buf.WriteString(conf.Order_Mac)
					buf.WriteString(mac)
				}, nil)
			}
		}
	} else {
		p := &safe.Strings{}
		p.Add(id)
		g_mac_order.Store(mac, p)
	}
}
func DelMacOrder(mac, id string) {
	if ids, ok := g_mac_order.Load(mac); ok {
		if i := ids.(*safe.Strings).Index(id); i >= 0 {
			ids.(*safe.Strings).Del(i)
		}
	}
}
func ClearMacOrder() {
	g_mac_order = sync.Map{}
}
