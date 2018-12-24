package logic

import (
	"common"
	"common/format"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"net/http"
	"netConfig"
	"time"
)

func Http_ask_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	passwd := q.Get("passwd")

	//FIXME:zhoumf: 限制同ip调用频率

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

func Http_timestamp(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%d", time.Now().Unix())))
}
