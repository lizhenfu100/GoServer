package main

import (
	"encoding/json"
	"net/http"
)

// ------------------------------------------------------------
// -- 登录服列表  http://www.chillyroom.com/api/
type TLogins struct {
	GameName string
	Logins   map[string][]string //<大区名, 地址>
}

var G_Logins map[string]*TLogins = nil //<游戏名, 登录列表>

func Http_get_login_list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v, ok := G_Logins[q.Get("name")]; ok {
		str, _ := json.MarshalIndent(v.Logins, "", "     ")
		w.Write(str)
	}
}
