package iqiyi

import (
	"crypto/md5"
	"fmt"
)

const (
	k_iqiyi = "iqiyi_Vhc#0b&OxAcuy"
)

//充值
type RechargeSuccess_req struct {
	User_id  string `json:"user_id"`  //用户id
	Role_id  string `json:"role_id"`  //角色id，没有传空
	Order_id string `json:"order_id"` //平台订单id，即 Third_order_id
	Money    int    `json:"money"`    //
	Time     int64  `json:"time"`     //时间戳
	UserData string `json:"userData"` //回传参数，存了游戏订单id
	Sign     string `json:"sign"`     //
}
type RechargeSuccess_ack struct {
	Result  int    `json:"result"`  //错误码：0 success, -1 sign error, -2 parameters error, -3 user not exists, -4 order repeat, -5 no server, -6 other errors
	Message string `json:"message"` //
}

//代金券订单查询
type VoucherOrder_req struct {
	Game_id    string `json:"game_id"`    //游戏id
	Start_time string `json:"start_time"` //查询开始时间
	End_time   string `json:"end_time"`   //查询结束时间，时间间隔不能超过31天
	//Order_type string `json:"order_type"` //1:完全抵扣订单、2:部分抵扣订单，默认为全部使用了代金券的订单
}
type VoucherOrder_ack struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Game_id     int     `json:"game_id"`     //游戏id
		Order_code  int     `json:"order_code"`  //订单号
		Order_money float32 `json:"order_money"` //用户下单金额
		Pay_money   float32 `json:"pay_money"`   //实际支付金额
		Pay_time    string  `json:"pay_time"`    //
		Send_time   string  `json:"send_time"`   //游戏币发送成功时间
		User_data   string  `json:"user_data"`   //CP回传数据
	} `json:"data"` //只有retcode是0才会有值，json
}

// -------------------------------------
//
func (self *RechargeSuccess_req) calcSign() string {
	s := fmt.Sprintf("%s%s%s%d%d%s", self.User_id, self.Role_id, self.Order_id, self.Money, self.Time, k_iqiyi)
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
func (self *VoucherOrder_req) calcSign() string {
	s := fmt.Sprintf("%d%s%s%s", self.Game_id, self.Start_time, self.End_time, k_iqiyi)
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
