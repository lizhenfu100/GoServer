/***********************************************************************
* @ 礼包码
* @ brief

* @ 接口文档
	· Rpc_login_get_gift
	· 上行参数
		· string key        礼包码key
		· uint32 pid        玩家playerId，可hash(uuid)代替
		· string pf_id      平台名，有些礼包仅固定平台领取
		· string version    客户端版本号，小于礼包版本，无法领
	· 下行参数
		· uint16 errCode
		· string json       客户端自行解析

* @ author zhoumf
* @ date 2018-12-12
***********************************************************************/
package logic

import (
	"bytes"
	"common/file"
	"net/http"
	"shared_svr/svr_login/gift_bag"
	"strconv"
)

func Http_gift_code_spawn(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key := q.Get("key")
	game := q.Get("game")
	count, _ := strconv.Atoi(q.Get("count"))

	gift_bag.MakeGiftCodeCsv(game, key, count)

	if q.Get("all") == "" {
		r.URL.Path = gift_bag.KGiftCodeDir + game + "/" + gift_bag.KGiftCodeTemp //只给新礼包码
	} else {
		r.URL.Path = gift_bag.KGiftCodeDir + game + "/" + key + ".csv" //历史所有礼包码
	}
	if count < 100 {
		var buf bytes.Buffer
		vs, _ := file.ReadCsv(r.URL.Path)
		for _, v := range vs {
			buf.WriteString(v[0])
			buf.WriteString("\n")
		}
		w.Write(buf.Bytes())
	} else {
		http.FileServer(http.Dir(".")).ServeHTTP(w, r)
	}
}
