package logic

import (
	"common"
	"common/format"
	"generate_out/err"
	"generate_out/rpc/enum"
	"net/http"
	"netConfig"
)

func Http_ask_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	passwd := q.Get("passwd")

	//! 创建回复
	ack := "Error: unknown"
	defer func() {
		w.Write([]byte(ack))
	}()

	if !format.CheckPasswd(passwd) {
		ack = "Error: Passwd_format_err"
		return
	}

	svrId := netConfig.HashCenterID(name)
	netConfig.CallRpcCenter(svrId, enum.Rpc_center_ask_reset_password, func(buf *common.NetPack) {
		buf.WriteString(name)
		buf.WriteString(passwd)
	}, func(recvBuf *common.NetPack) {
		errCode := recvBuf.ReadUInt16()
		switch errCode {
		case err.Success:
			ack = "Ok: send verification to your email. Please check it."
		case err.Account_none:
			ack = "Error: Account_none"
		case err.Account_without_bind_info:
			ack = "Error: Account_without_bind_info"
		}
	})
}
