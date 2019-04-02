package main

import (
	"common"
	"common/std/sign"
	"conf"
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
	accountName := q.Get("name")

	if accountName == "" {
		w.Write(common.S2B("AccountName nil"))
	} else {
		mhttp.CallRpc(g_addrs.CenterAddr, enum.Rpc_center_player_login_addr, func(buf *common.NetPack) {
			buf.WriteString(conf.GameName)
			buf.WriteString(accountName)
		}, func(recvBuf *common.NetPack) {
			if e := recvBuf.ReadUInt16(); e == err.Success {
				loginIp := recvBuf.ReadString()
				loginPort := recvBuf.ReadUInt16()
				gameIp := recvBuf.ReadString()
				gamePort := recvBuf.ReadUInt16()

				w.Write(common.S2B(fmt.Sprintf("LoginAddr: %s:%d\nGameAddr: %s:%d",
					loginIp, loginPort, gameIp, gamePort)))
			}
		})
	}
}
func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//1、创建url
	u, _ := url.Parse(g_addrs.CenterAddr + "/reset_password")
	//2、写入参数
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(q.Get("pwd")+flag))
	//3、生成完整url
	u.RawQuery = q.Encode()
	if buf := mhttp.Get(u.String()); buf != nil {
		w.Write(buf)
	}
}
func Http_relay_to_save(w http.ResponseWriter, r *http.Request) {
	for _, v := range g_addrs.SaveAddrs {
		u, _ := url.Parse(v + r.RequestURI) //除去域名或ip的url
		if buf := mhttp.Get(u.String()); buf != nil {
			w.Write(buf)
		}
	}
}
