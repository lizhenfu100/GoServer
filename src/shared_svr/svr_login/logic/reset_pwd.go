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
	"netConfig/meta"
	mhttp "nets/http/http"
	"shared_svr/svr_login/logic/cache"
	"strconv"
	"strings"
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
	} else if pMeta := meta.GetMeta("center", netConfig.HashCenterID(v)); pMeta != nil {
		centerAddr := fmt.Sprintf("http://%s:%d", pMeta.OutIP, pMeta.HttpPort)
		//1、创建url
		u, _ := url.Parse(centerAddr + "/reset_password")
		q2 := u.Query()
		//2、写入参数
		q2.Set("k", k)
		q2.Set("v", v)
		q2.Set("pwd", passwd)
		flag := strconv.FormatInt(time.Now().Unix(), 10)
		q2.Set("flag", flag)
		q2.Set("language", language)
		q2.Set("sign", sign.CalcSign(k+v+passwd+flag))
		//3、生成完整url
		u.RawQuery = q2.Encode()
		switch cache.Del(v); k {
		case "email":
			errCode = email.SendMail("Reset Password", v, u.String(), language)
		case "phone":
			if code := q.Get("code"); sms.CheckCode(v, code) {
				if r, e := http.Get(u.String()); e == nil {
					if strings.Index(common.B2S(mhttp.ReadBody(r.Body)), "ok") < 0 {
						errCode = err.Not_found
					} else {
						errCode = err.Success
					}
				}
			} else {
				errCode = err.Token_verify_err
			}
		}
	}
}
func Rpc_login_sms_code(req, ack *common.NetPack, _ common.Conn) { //发手机验证码
	phone := req.ReadString()
	if format.CheckBindValue("phone", phone) {
		ack.WriteUInt16(sms.SendCode(phone))
	} else {
		ack.WriteUInt16(err.Phone_format_err)
	}
}
func Rpc_login_sms_check(req, ack *common.NetPack, _ common.Conn) {
	phone := req.ReadString()
	code := req.ReadString()
	ack.WriteBool(sms.CheckCode(phone, code))
}

// ------------------------------------------------------------
func Http_timestamp(w http.ResponseWriter, r *http.Request) {
	w.Write(common.S2B(fmt.Sprintf("%d", time.Now().Unix())))
}
