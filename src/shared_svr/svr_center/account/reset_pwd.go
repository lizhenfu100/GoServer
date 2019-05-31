package account

import (
	"common"
	"common/format"
	"common/std/sign"
	"common/tool/email"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	emailAddr := q.Get("email")
	passwd := q.Get("pwd")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

	//! 创建回复
	ack := "Error: unknown"
	defer func() {
		w.Write(common.S2B(ack))
	}()

	if sign.CalcSign(passwd+flag) != q.Get("sign") {
		ack, _ = email.Translate("Error: sign failed", "")
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack, _ = email.Translate("Error: url expire", "")
	} else if !format.CheckPasswd(passwd) {
		ack, _ = email.Translate("Error: Passwd_format_err", "")
	} else if account := GetAccountByBindInfo("email", emailAddr); account == nil {
		ack, _ = email.Translate("Error: Account_none", "")
	} else {
		account.SetPasswd(passwd)
		dbmgo.UpdateId(KDBTable, account.AccountID, bson.M{"$set": bson.M{
			"password": account.Password}})
		ack, _ = email.Translate("Reset password ok", "")
	}
}
