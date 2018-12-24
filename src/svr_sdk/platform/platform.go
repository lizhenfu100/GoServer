package platform

import (
	"svr_sdk/msg"
	"svr_sdk/platform/kuaishou"
	"svr_sdk/platform/x7sy"
)

//客户端请求生成订单【各渠道可能回复数据不同】
func NewPreBuyAck(pf_id string) msg.Pre_buy_ack {
	switch pf_id {
	case "kuaishou":
		return &kuaishou.Pre_buy_ack{}
	case "xiao7":
		return &x7sy.Pre_buy_ack{}
	default:
		return &msg.Pre_buy{}
	}
}
