package logic

import (
	"common"
	"common/format"
	"common/std/sign"
	"common/tool/email"
	"common/tool/sms"
	"encoding/binary"
	"fmt"
	"generate_out/err"
	"net/http"
	"net/url"
	"netConfig"
	"shared_svr/svr_login/logic/cache"
	"strconv"
	"time"
)

func Http_ask_reset_password(w http.ResponseWriter, r *http.Request) { //提示查收邮件
	q := r.URL.Query()
	k, v := q.Get("key"), q.Get("name")
	passwd := q.Get("passwd")
	language := q.Get("language")
	sign.Decode(&passwd)
	if k == "" {
		k = "email" //TODO:待删除，兼容老客户端
	}
	errCode := err.Unknow_error
	defer func() {
		ack := make([]byte, 2)
		binary.LittleEndian.PutUint16(ack, errCode)
		w.Write(ack)
	}()
	if !format.CheckPasswd(passwd) {
		errCode = err.Passwd_format_err
	} else {
		centerAddr := netConfig.GetHttpAddr("center", netConfig.HashCenterID(v))
		//1、创建url
		u, _ := url.Parse(centerAddr + "/reset_password")
		q := u.Query()
		//2、写入参数
		q.Set("k", k)
		q.Set("v", v)
		q.Set("pwd", passwd)
		flag := strconv.FormatInt(time.Now().Unix(), 10)
		q.Set("flag", flag)
		q.Set("language", language)
		q.Set("sign", sign.CalcSign(k+v+passwd+flag))
		//3、生成完整url
		u.RawQuery = q.Encode()
		switch cache.Del(v); k {
		case "email":
			errCode = email.SendMail("Reset Password", v, u.String(), language)
		case "phone":
			if code := q.Get("code"); sms.CheckCode(v, code) {
				if _, e := http.Get(u.String()); e == nil {
					errCode = err.Success
				}
			} else {
				errCode = err.Token_verify_err
			}
		}
	}
}
func Rpc_login_sms_code(req, ack *common.NetPack) { //发手机验证码
	phone := req.ReadString()
	if format.CheckBindValue("phone", phone) {
		sms.SendCode(phone)
	}
}
func Rpc_login_sms_check(req, ack *common.NetPack) {
	phone := req.ReadString()
	code := req.ReadString()
	ack.WriteBool(sms.CheckCode(phone, code))
}

// ------------------------------------------------------------
func Http_timestamp(w http.ResponseWriter, r *http.Request) {
	w.Write(common.S2B(fmt.Sprintf("%d", time.Now().Unix())))
}
