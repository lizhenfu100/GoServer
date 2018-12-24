package account

import (
	"common"
	"common/email"
	"common/format"
	"common/sign"
	"dbmgo"
	"fmt"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"netConfig/meta"
	"strconv"
	"time"
)

func Rpc_center_ask_reset_password(req, ack *common.NetPack) {
	name := req.ReadString()
	passwd := req.ReadString()

	if !format.CheckPasswd(passwd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if account := GetAccountByName(name); account == nil {
		ack.WriteUInt16(err.Account_none)
	} else if emailAddr, ok := account.BindInfo["email"]; !ok {
		ack.WriteUInt16(err.Account_without_bind_info)
	} else {
		ack.WriteUInt16(err.Success)

		//1、创建url
		httpAddr := fmt.Sprintf("http://%s:%d/reset_password",
			meta.G_Local.OutIP, meta.G_Local.Port())
		u, _ := url.Parse(httpAddr)
		q := u.Query()
		//2、写入参数
		q.Set("id", strconv.Itoa(int(account.AccountID)))
		q.Set("pwd", passwd)
		flag := strconv.FormatInt(time.Now().Unix(), 10)
		q.Set("flag", flag)
		q.Set("sign", sign.CalcSign(passwd+flag))
		//3、生成完整url
		u.RawQuery = q.Encode()
		email.SendMail("密码重置", emailAddr, u.String())
	}
}
func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	aid, _ := strconv.Atoi(q.Get("id"))
	passwd := q.Get("pwd")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 0)

	//FIXME:zhoumf: 限制同ip调用频率

	//! 创建回复
	ack := "Error: unknown"
	defer func() {
		w.Write([]byte(ack))
	}()

	if sign.CalcSign(passwd+flag) != q.Get("sign") {
		ack = "Error: sign failed"
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack = "Error: url expire"
	} else if !format.CheckPasswd(passwd) {
		ack = "Error: Passwd_format_err"
	} else if account := GetAccountById(uint32(aid)); account == nil {
		ack = "Error: Account_none"
	} else {
		account.SetPasswd(passwd)
		dbmgo.UpdateId(kDBTable, account.AccountID, bson.M{"$set": bson.M{
			"password": passwd}})
		ack = "Reset password ok"
	}
}
