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

// ------------------------------------------------------------
type TLogins struct {
	Game string
	IPs  map[string][]string //<大区, 地址>
}

var G_Logins map[string]*TLogins = nil //<游戏名, 大区列表>

func Http_svr_list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v, ok := G_Logins[q.Get("game")]; ok {
		if v, ok := v.IPs[q.Get("region")]; ok {
			b, _ := json.Marshal(v)
			w.Write(b)
		}
	}
}

// ------------------------------------------------------------
