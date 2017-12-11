package huluxia

import (
	"bytes"
	"crypto/md5"
	"fmt"
)

const (
	k_huluxia = "91d15a277f603490129317f874c7bae6"
)

//充值
type RechargeSuccess_req struct {
	Sign_type    string  `json:"sign_type"`    //签名方式，固定为"MD5"
	Sign         string  `json:"sign"`         //
	Out_order_no string  `json:"out_order_no"` //游戏订单id
	Subject      string  `json:"subject"`      //商品名称
	Order_no     string  `json:"order_no"`     //平台订单订单id，即 Third_order_id
	Order_status string  `json:"order_status"` //订单状态
	Apk_id       string  `json:"apk_id"`       //应用ID
	Uid          string  `json:"uid"`          //用户ID
	Total_amount float32 `json:"total_amount"` //总金额
}

// -------------------------------------
//
func (self *RechargeSuccess_req) CalcSign() string {
	var buf bytes.Buffer
	//排除空字段，Sign_type，Sign
	if self.Apk_id != "" {
		buf.WriteString("apk_id=")
		buf.WriteString(self.Apk_id)
		buf.WriteString("&")
	}
	buf.WriteString("order_no=")
	buf.WriteString(self.Order_no)
	buf.WriteString("&order_status=")
	buf.WriteString(self.Order_status)
	buf.WriteString("&out_order_no=")
	buf.WriteString(self.Out_order_no)
	buf.WriteString("&subject=")
	buf.WriteString(self.Subject)
	buf.WriteString("&total_amount=")
	buf.WriteString(fmt.Sprintf("%.2f", self.Total_amount))
	buf.WriteString("&uid=")
	buf.WriteString(self.Uid)

	buf.WriteString(k_huluxia)
	s := buf.String()
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
