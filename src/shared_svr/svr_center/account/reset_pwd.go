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
	k, v := q.Get("k"), q.Get("v")
	passwd := q.Get("pwd")
	flag := q.Get("flag")
	language := q.Get("language")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

	ack := "Error: unknown"
	defer func() {
		w.Write(common.S2B(ack))
	}()
	if sign.CalcSign(k+v+passwd+flag) != q.Get("sign") {
		ack, _ = email.Translate("Error: sign failed", language)
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack, _ = email.Translate("Error: url expire", language)
	} else if !format.CheckPasswd(passwd) {
		ack, _ = email.Translate("Error: Passwd_format_err", language)
	} else if p := GetAccountByBindInfo(k, v); p == nil {
		ack, _ = email.Translate("Error: Account_none", language)
	} else {
		p.SetPasswd(passwd)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"password": p.Password}})
		p.verifyEmailOK()
		ack, _ = email.Translate("Reset password ok", language)
	}
}
