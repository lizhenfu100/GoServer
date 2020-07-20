/***********************************************************************
* @ DNSpod

* @ 解析优先级
	· 自定义线路 > 分省线路 > 分区域线路 > 分运营商线路 > 分国家线路 > 分大洲线路 > 默认线路

* @ 查ip
	· nslookup game.chillyroom.com

* @ author zhoumf
* @ date 2020-4-3
***********************************************************************/
package logic

import (
	"encoding/json"
	"net/http"
)

// http://game.chillyroom.com:7233/svr_list?game=HappyDiner

// ------------------------------------------------------------
//go:generate D:\server\bin\gen_conf.exe logic logins
type logins map[string]*struct {
	Game string
	IPs  map[string][]string
	//大区名：America、Asia、Europe、Brazil、ChinaNorth、ChinaSouth、MiddleEast、Australia
	//服务名：file、nric
}

func Http_svr_list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v, ok := Logins()[q.Get("game")]; ok {
		var ips []string
		if svr := q.Get("svr"); svr != "" {
			ips = v.IPs[svr]
		} else if ips, ok = v.IPs[q.Get("region")]; !ok {
			ips = v.IPs[""]
		}
		b, _ := json.Marshal(ips)
		w.Write(b)
	}
}

// ------------------------------------------------------------
