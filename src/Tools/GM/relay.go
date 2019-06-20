package main

import (
	"common"
	"common/std/sign"
	"generate_out/rpc/enum"
	"net/http"
	"net/url"
	mhttp "nets/http"
	"strconv"
	"time"
)

func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//1、创建url
	u, _ := url.Parse(g_common.CenterAddr + "/reset_password")
	//2、写入参数
	pwd := q.Get("pwd")
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(pwd+flag))
	//3、生成完整url
	u.RawQuery = q.Encode()
	if buf := mhttp.Client.Get(u.String()); buf != nil {
		w.Write(buf)
	}
}
func Http_bind_info_force(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//1、创建url
	u, _ := url.Parse(g_common.CenterAddr + "/bind_info_force")
	//2、写入参数
	name, k, v := q.Get("name"), q.Get("k"), q.Get("v")
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(name+k+v+flag))
	//3、生成完整url
	u.RawQuery = q.Encode()
	if buf := mhttp.Client.Get(u.String()); buf != nil {
		w.Write(buf)
	}
}

func Http_relay_to_save(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gameName := q.Get("game")

	for i := 0; i < len(g_list); i++ {
		if p := &g_list[i]; p.GameName == gameName {
			for _, v := range p.SaveAddrs {
				u, _ := url.Parse(v + r.RequestURI) //除去域名或ip的url
				if buf := mhttp.Client.Get(u.String()); buf != nil {
					w.Write(buf)
				}
			}
			return
		}
	}
	w.Write(common.S2B("GameName nil"))
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
