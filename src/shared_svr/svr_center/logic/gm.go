package logic

import (
	"common"
	"encoding/json"
	"net/http"
	"shared_svr/svr_center/account"
)

func Http_show_account_info(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if p := account.GetAccountByName(q.Get("name")); p != nil {
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	} else {
		w.Write(common.S2B("none account"))
	}
}
