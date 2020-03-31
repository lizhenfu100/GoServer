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
		if k == "email" {
			ack, _ = email.Translate(ack, language)
		}
		w.Write(common.S2B(ack))
	}()
	if sign.CalcSign(k+v+passwd+flag) != q.Get("sign") {
		ack = "Error: sign failed"
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack = "Error: url expire"
	} else if !format.CheckPasswd(passwd) {
		ack = "Error: Passwd_format_err"
	} else if _, p := GetAccountByBindInfo(k, v); p == nil {
		ack = "Error: Account_none"
	} else {
		ack = "Reset password ok"
		p.SetPasswd(passwd)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"password": p.Password}})
		if k == "email" {
			p.verifyEmailOK()
		}
		CacheDel(p)
	}
}
