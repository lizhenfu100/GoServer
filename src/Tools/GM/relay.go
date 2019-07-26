package main

import (
	"common"
	"common/std/sign"
	"generate_out/rpc/enum"
	"net/http"
	"net/url"
	mhttp "nets/http"
	"shared_svr/svr_login/gm"
	"strconv"
	"time"
)

func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//1、追加参数
	k, v := q.Get("k"), q.Get("v")
	pwd := q.Get("pwd")
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(k+v+pwd+flag))
	for i := 0; i < len(g_common.CenterList); i++ {
		//2、创建url
		u, _ := url.Parse(g_common.CenterList[i] + "/reset_password")
		//3、生成完整url
		u.RawQuery = q.Encode()
		if buf := mhttp.Client.Get(u.String()); buf != nil && i == 0 {
			w.Write(buf)
		}
	}
}
func Http_bind_info_force(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//1、追加参数
	aid, k, v := q.Get("aid"), q.Get("k"), q.Get("v")
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(aid+k+v+flag))
	for i := 0; i < len(g_common.CenterList); i++ {
		//2、创建url
		u, _ := url.Parse(g_common.CenterList[i] + "/bind_info_force")
		//3、生成完整url
		u.RawQuery = q.Encode()
		if buf := mhttp.Client.Get(u.String()); buf != nil && i == 0 {
			w.Write(buf)
		}
	}
}

func Http_relay_gm_cmd(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	addr := q.Get("addr")
	cmd := q.Get("cmd")

	mhttp.CallRpc("http://"+addr, enum.Rpc_gm_cmd, func(buf *common.NetPack) {
		buf.WriteString(cmd)
	}, func(recvBuf *common.NetPack) {
		str := recvBuf.ReadString()
		w.Write(common.S2B(str))
	})
}

// ------------------------------------------------------------
// 批量转发
func Http_relay_to_save(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gameName := q.Get("game")

	if p, ok := g_map[gameName]; !ok {
		w.Write(common.S2B("GameName error"))
	} else {
		var acks [][]byte
		for _, v := range p.Logins {
			for _, v2 := range v.Games {
				for _, v3 := range v2.Saves {
					u, _ := url.Parse(v3 + r.RequestURI) //除去域名或ip的url
					if buf := mhttp.Client.Get(u.String()); buf != nil {
						acks = append(acks, buf)
					}
				}
			}
		}
		writeRelayResult(w, acks)
	}
}
func Http_relay_to_login(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gameName := q.Get("game")

	if p, ok := g_map[gameName]; !ok {
		w.Write(common.S2B("GameName error"))
	} else {
		var acks [][]byte
		for _, v := range p.Logins {
			u, _ := url.Parse(v.Login + r.RequestURI) //除去域名或ip的url
			if buf := mhttp.Client.Get(u.String()); buf != nil {
				acks = append(acks, buf)
			}
		}
		writeRelayResult(w, acks)
	}
}
func writeRelayResult(w http.ResponseWriter, acks [][]byte) {
	//检查回复是否一致
	isSame := true
	for i := 0; i < len(acks); i++ {
		for j := i + 1; j < len(acks); j++ {
			if common.B2S(acks[i]) != common.B2S(acks[j]) {
				isSame = false
			}
		}
	}
	if isSame {
		w.Write(acks[0])
	} else {
		w.Write(common.S2B("Different result !!!\n\n"))
		for i := 0; i < len(acks); i++ {
			w.Write(acks[i])
		}
	}
}

// ------------------------------------------------------------
// 生成礼包码
func Http_gift_code_spawn(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key := q.Get("key")
	count, _ := strconv.Atoi(q.Get("count"))

	gm.MakeGiftCodeCsv(key, count)

	r.URL.Path = gm.KGiftCodeDir + key + ".csv"
	svr := http.FileServer(http.Dir("."))
	svr.ServeHTTP(w, r)
}
