package logic

import (
	"common"
	"common/email"
	"common/sign"
	"encoding/binary"
	"generate_out/err"
	"generate_out/rpc/enum"
	mhttp "http"
	"net/http"
	"net/url"
	"netConfig"
	"strconv"
	"time"
)

// Client须提示玩家查收邮件
func Http_ask_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	passwd := q.Get("passwd")

	//! 创建回复
	errCode := err.Unknow_error
	defer func() {
		ack := make([]byte, 2)
		binary.LittleEndian.PutUint16(ack, errCode)
		w.Write(ack)
	}()

	centerAddr := netConfig.GetHttpAddr("center", netConfig.HashCenterID(name))
	if centerAddr == "" {
		errCode = err.None_center_server
		return
	}
	mhttp.CallRpc(centerAddr, enum.Rpc_center_ask_reset_password, func(buf *common.NetPack) {
		buf.WriteString(name)
		buf.WriteString(passwd)
	}, func(recvBuf *common.NetPack) {
		if errCode = recvBuf.ReadUInt16(); errCode == err.Success {
			accountId := recvBuf.ReadUInt32()
			emailAddr := recvBuf.ReadString()
			//1、创建url
			u, _ := url.Parse(centerAddr + "/reset_password")
			q := u.Query()
			//2、写入参数
			q.Set("id", strconv.FormatInt(int64(accountId), 10))
			q.Set("pwd", passwd)
			flag := strconv.FormatInt(time.Now().Unix(), 10)
			q.Set("flag", flag)
			q.Set("sign", sign.CalcSign(passwd+flag))
			//3、生成完整url
			u.RawQuery = q.Encode()
			body := email.CreateTemplate("csv/email/reset_pwd.txt", u.String())
			email.SendMail("Reset Password", emailAddr, body)
		}
	})
}
