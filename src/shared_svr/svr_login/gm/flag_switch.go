package gm

import (
	"common"
	"common/timer"
	"conf"
	"dbmgo"
	"net/http"
	"shared_svr/svr_login/logic"
	"strconv"
)

func Http_flag_switch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	falg := q.Get("flag")
	val := q.Get("v")

	switch falg {
	case "banLogin":
		if val == "0" {
			logic.G_banLogin = false
			timer.AddTimer(func() {
				logic.G_banLogin = true
			}, 3600*2, 0, 0)
		} else {
			logic.G_banLogin = true
		}
	default:
		w.Write([]byte("fail"))
		return
	}
	w.Write(common.S2B(falg + " " + val))
}

// ------------------------------------------------------------
// 运营用自增id
func Http_get_inc_id(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if key := q.Get("key"); key != "" {
		id := dbmgo.GetNextIncId(key)
		w.Write(common.S2B(strconv.FormatInt(int64(id), 10)))
	}
}
