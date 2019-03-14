package logic

import (
	"conf"
	"dbmgo"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"shared_svr/svr_center/account"
)

// 解封账号
func Http_permit_account(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	v := q.Get("val")

	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	if p := account.GetAccountByName(v); p != nil {
		p.IsForbidden = false
		dbmgo.UpdateId(account.KDBTable, p.AccountID, bson.M{"$set": bson.M{
			"isforbidden": false}})
		w.Write([]byte("ok"))
	} else {
		w.Write([]byte("none account"))
	}
}

func Http_show_account_info(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")

	if p := account.GetAccountByName(name); p != nil {
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	} else {
		w.Write([]byte("none account"))
	}
}
