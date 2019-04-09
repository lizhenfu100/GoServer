package logic

import (
	"common"
	"common/format"
	"common/std/sign"
	"common/tool/email"
	"encoding/binary"
	"generate_out/err"
	"generate_out/rpc/enum"
	"net/http"
	"net/url"
	"netConfig"
	mhttp "nets/http"
	"strconv"
	"time"
)

// Client须提示玩家查收邮件
func Http_ask_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	passwd := q.Get("passwd")
	language := q.Get("language")

	//! 创建回复
	errCode := err.Unknow_error
	defer func() {
		ack := make([]byte, 2)
		binary.LittleEndian.PutUint16(ack, errCode)
		w.Write(ack)
	}()

	centerId := netConfig.HashCenterID(name)

	if !format.CheckPasswd(passwd) {
		errCode = err.Passwd_format_err
	} else if centerAddr := netConfig.GetHttpAddr("center", centerId); centerAddr == "" {
		errCode = err.None_center_server
	} else {
		mhttp.CallRpc(centerAddr, enum.Rpc_center_get_bind_info, func(buf *common.NetPack) {
			buf.WriteString(name)
			buf.WriteString("email")
		}, func(recvBuf *common.NetPack) {
			if emailAddr := recvBuf.ReadString(); emailAddr == "" {
				errCode = err.Account_without_bind_info
			} else {
				//1、创建url
				u, _ := url.Parse(centerAddr + "/reset_password")
				q := u.Query()
				//2、写入参数
				q.Set("name", name)
				q.Set("pwd", passwd)
				flag := strconv.FormatInt(time.Now().Unix(), 10)
				q.Set("flag", flag)
				q.Set("sign", sign.CalcSign(passwd+flag))
				//3、生成完整url
				u.RawQuery = q.Encode()
				email.SendMail("Reset Password", emailAddr, u.String(), language)
				errCode = err.Success
			}
		})
	}
}
