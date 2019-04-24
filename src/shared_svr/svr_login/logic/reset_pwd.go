package logic

import (
	"common/format"
	"common/std/sign"
	"common/tool/email"
	"encoding/binary"
	"gamelog"
	"generate_out/err"
	"net/http"
	"net/url"
	"netConfig"
	"strconv"
	"time"
)

// Client须提示玩家查收邮件
func Http_ask_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	emailAddr := q.Get("name")
	passwd := q.Get("passwd")
	language := q.Get("language")
	gamelog.Debug("ask_reset_password: %s %s %s", emailAddr, passwd, language)

	//! 创建回复
	errCode := err.Unknow_error
	defer func() {
		ack := make([]byte, 2)
		binary.LittleEndian.PutUint16(ack, errCode)
		w.Write(ack)
	}()

	if !format.CheckPasswd(passwd) {
		errCode = err.Passwd_format_err
	} else {
		centerAddr := netConfig.GetHttpAddr("center", netConfig.HashCenterID(emailAddr))
		//1、创建url
		u, _ := url.Parse(centerAddr + "/reset_password")
		q := u.Query()
		//2、写入参数
		q.Set("email", emailAddr)
		q.Set("pwd", passwd)
		flag := strconv.FormatInt(time.Now().Unix(), 10)
		q.Set("flag", flag)
		q.Set("sign", sign.CalcSign(passwd+flag))
		//3、生成完整url
		u.RawQuery = q.Encode()
		email.SendMail2("Reset Password", emailAddr, u.String(), language)
		errCode = err.Success
	}
}
