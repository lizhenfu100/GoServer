package main

import (
	"common"
	"conf"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	mhttp "http"
	"net/http"
)

func Http_query_account_login_addr(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	accountName := q.Get("name")

	if accountName == "" {
		w.Write([]byte("AccountName nil"))
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

				w.Write([]byte(fmt.Sprintf("LoginAddr: %s:%d\nGameAddr: %s:%d",
					loginIp, loginPort, gameIp, gamePort)))
			}
		})
	}
}
