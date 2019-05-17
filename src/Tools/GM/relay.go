package main

import (
	"common"
	"common/std/sign"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"net/http"
	"net/url"
	mhttp "nets/http"
	"strconv"
	"time"
)

func Http_query_account_login_addr(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gameName := q.Get("game")
	accountName := q.Get("name")

	if accountName == "" {
		w.Write(common.S2B("AccountName nil"))
	} else {
		mhttp.CallRpc(g_common.CenterAddr, enum.Rpc_center_player_login_addr, func(buf *common.NetPack) {
			buf.WriteString(gameName)
			buf.WriteString(accountName)
		}, func(recvBuf *common.NetPack) {
			if e := recvBuf.ReadUInt16(); e == err.Success {
				loginIp := recvBuf.ReadString()
				loginPort := recvBuf.ReadUInt16()
				gameIp := recvBuf.ReadString()
				gamePort := recvBuf.ReadUInt16()

				w.Write(common.S2B(fmt.Sprintf("LoginAddr: %s:%d\nGameAddr: %s:%d",
					loginIp, loginPort, gameIp, gamePort)))
			} else if e == err.Svr_not_working {
				w.Write(common.S2B("Svr_not_working"))
			} else if e == err.Not_found {
				w.Write(common.S2B(gameName + ": none data"))
			} else {
				w.Write(common.S2B("Unknow_error"))
			}
		})
	}
}
func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//1、创建url
	u, _ := url.Parse(g_common.CenterAddr + "/reset_password")
	//2、写入参数
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(q.Get("pwd")+flag))
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
