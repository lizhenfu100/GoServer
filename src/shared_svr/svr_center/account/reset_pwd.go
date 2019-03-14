package account

import (
	"common"
	"common/format"
	"common/sign"
	"dbmgo"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"net/http"
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
		ack.WriteUInt32(account.AccountID)
		ack.WriteString(emailAddr)
	}
}
func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	aid, _ := strconv.Atoi(q.Get("id"))
	passwd := q.Get("pwd")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

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
		dbmgo.UpdateId(KDBTable, account.AccountID, bson.M{"$set": bson.M{
			"password": account.Password}})
		ack = "Reset password ok"
	}
}
