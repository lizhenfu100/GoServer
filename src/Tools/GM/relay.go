package main

import (
	"common"
	"common/std/sign"
	"conf"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	mhttp "http"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func Http_query_account_login_addr(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	accountName := q.Get("name")

	if accountName == "" {
		w.Write(common.S2B("AccountName nil"))
	} else {
		mhttp.CallRpc(g_templateData.CenterAddr, enum.Rpc_center_player_login_addr, func(buf *common.NetPack) {
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
	q1 := r.URL.Query()
	aid := q1.Get("id")
	passwd := q1.Get("pwd")

	//1、创建url
	u, _ := url.Parse(g_templateData.CenterAddr + "/reset_password")
	q := u.Query()
	//2、写入参数
	q.Set("id", aid)
	q.Set("pwd", passwd)
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(passwd+flag))
	//3、生成完整url
	u.RawQuery = q.Encode()
	if buf := mhttp.Get(u.String()); buf != nil {
		w.Write(buf)
	}
}
