/***********************************************************************
* @ sdk是平台性质的，各游戏通用
	、既支持纯单机业务；也可后台下单，到账通知svr_game

* @ 不同平台下的同一种支付渠道pay_id必须一样
	iapppay		Pay_id:1
	iqiyi		Pay_id:8
	huluxia		Pay_id:9 	平台名Pf_id:gourd
	kuaishou	Pay_id:10
	x7sy		Pay_id:11
	kuaishou2   Pay_id:12
	pingxx      Pay_id:100
	xdpublic	Pay_id:101
	midas		Pay_id:102
***********************************************************************/
package platform

import (
	"fmt"
	"net/http"
	"shared_svr/svr_sdk/msg"
	"shared_svr/svr_sdk/platform/kuaishou"
	"shared_svr/svr_sdk/platform/kuaishou2"
	"shared_svr/svr_sdk/platform/midas"
	"shared_svr/svr_sdk/platform/pingxx"
	"shared_svr/svr_sdk/platform/x7sy"
	"shared_svr/svr_sdk/platform/xdpublic"
)

const (
	k_svr_crt = "rsa/server.crt" //服务端的数字证书文件路径
	k_svr_key = "rsa/server.key" //服务端的私钥文件路径
)

func Init() {
	go func() {
		_svr := http.Server{Addr: ":7000"}
		if e := _svr.ListenAndServeTLS(k_svr_crt, k_svr_key); e != nil {
			fmt.Println("https init: ", e.Error())
		}
	}()
	go func() {
		var _svr http.Server //TODO：待删除
		if e := _svr.ListenAndServeTLS(k_svr_crt, k_svr_key); e != nil {
			fmt.Println("midas init: ", e.Error())
		}
	}()
}

//客户端请求生成订单，各渠道数据不同
func NewPreBuyAck(pay_id int) msg.IPre_buy_ack {
	switch pay_id {
	case 10: //快手
		return &kuaishou.Pre_buy_ack{}
	case 12: //快手
		return &kuaishou2.Pre_buy_ack{}
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
