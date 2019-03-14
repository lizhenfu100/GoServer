/***********************************************************************
* @ sdk是平台性质的，各游戏通用
	、既支持纯单机业务；也可后台下单，到账通知svr_game

* @ 不同平台下的同一种支付渠道pay_id必须一样
	iapppay
		支付商 Pay_id：1
	iqiyi
		支付商 Pay_id：8
	huluxia
		支付商 Pay_id：9 平台名Pf_id:gourd
	kuaishou
		支付商 Pay_id：10
	x7sy
		支付商 Pay_id：11
***********************************************************************/
package platform

import (
	"shared_svr/svr_sdk/msg"
	"shared_svr/svr_sdk/platform/iapppay"
	"shared_svr/svr_sdk/platform/kuaishou"
	"shared_svr/svr_sdk/platform/x7sy"
)

//客户端请求生成订单【各渠道可能回复数据不同】
func NewPreBuyAck(pf_id string) msg.Pre_buy_ack {
	switch pf_id {
	case "origin": //官网包，走爱贝支付
		return &iapppay.Pre_buy_ack{}
	case "kuaishou":
		return &kuaishou.Pre_buy_ack{}
	case "xiao7":
		return &x7sy.Pre_buy_ack{}
	default:
		return &msg.Pre_buy{}
	}
}
