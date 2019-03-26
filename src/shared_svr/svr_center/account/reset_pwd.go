package account

import (
	"common"
	"common/format"
	"common/std/sign"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	passwd := q.Get("pwd")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

	//! 创建回复
	ack := "Error: unknown"
	defer func() {
		w.Write(common.S2B(ack))
	}()

	if sign.CalcSign(passwd+flag) != q.Get("sign") {
		ack = "Error: sign failed"
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack = "Error: url expire"
	} else if !format.CheckPasswd(passwd) {
		ack = "Error: Passwd_format_err"
	} else if account := GetAccountByName(name); account == nil {
		ack = "Error: Account_none"
	} else {
		account.SetPasswd(passwd)
		dbmgo.UpdateId(KDBTable, account.AccountID, bson.M{"$set": bson.M{
			"password": account.Password}})
		ack = "Reset password ok"
	}
}
