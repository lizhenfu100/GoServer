/***********************************************************************
* @ sdk是平台性质的，各游戏通用
	、既支持纯单机业务；也可后台下单，到账通知svr_game

* @ 不同平台下的同一种支付渠道pay_id必须一样
	iapppay		Pay_id:1
	iqiyi		Pay_id:8
	huluxia		Pay_id:9 	平台名Pf_id:gourd
	kuaishou	Pay_id:10
	x7sy		Pay_id:11
	pingxx      Pay_id:100
	xdpublic	Pay_id:101
	midas		Pay_id:102
***********************************************************************/
package platform

import (
	"shared_svr/svr_sdk/msg"
	"shared_svr/svr_sdk/platform/kuaishou"
	"shared_svr/svr_sdk/platform/midas"
	"shared_svr/svr_sdk/platform/pingxx"
	"shared_svr/svr_sdk/platform/x7sy"
	"shared_svr/svr_sdk/platform/xdpublic"
)

//客户端请求生成订单，各渠道数据不同
func NewPreBuyAck(pay_id int) msg.IPre_buy_ack {
	switch pay_id {
	case 10: //快手
		return &kuaishou.Pre_buy_ack{}
	case 11: //小七
		return &x7sy.Pre_buy_ack{}
	case 100: //Ping++
		return &pingxx.Pre_buy_ack{}
	case 101: //火树
		return &xdpublic.Pre_buy_ack{}
	case 102: //米大师
		return &midas.Pre_buy_ack{}
	default:
		return &msg.Pre_buy_ack{}
	}
}
func ConfirmOrder(order *msg.TOrderInfo) { //发货后的回调
	msg.ConfirmOrder(order)
	switch order.Pay_id {
	case 102: //米大师
		return //TODO:zhoumf:通知第三方后台
	}
}
func Init() {
	midas.Init()
}
