package logic

import (
	"common"
	"common/std/sign"
	"common/tool/email"
	"encoding/binary"
	"fmt"
	"generate_out/err"
	"net/http"
	"net/url"
	"netConfig"
	"netConfig/meta"
	"strconv"
	"time"
)

func Rpc_login_send_email(req, ack *common.NetPack, _ common.Conn) {
	subject := req.ReadString()
	target := req.ReadString()
	body := req.ReadString()
	language := req.ReadString()

	e := email.SendMail(subject, target, body, language)
	ack.WriteUInt16(e)
}

// Client须提示玩家查收邮件
func Http_verify_email(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	addr := q.Get("email")
	language := q.Get("language")

	//! 创建回复
	errCode := err.Unknow_error
	defer func() {
		ack := make([]byte, 2)
		binary.LittleEndian.PutUint16(ack, errCode)
		w.Write(ack)
	}()
	return
	if pMeta := meta.GetMeta("center", netConfig.HashCenterID(addr)); pMeta != nil {
		centerAddr := fmt.Sprintf("http://%s:%d", pMeta.OutIP, pMeta.HttpPort)
		//1、增加参数
		flag := strconv.FormatInt(time.Now().Unix(), 10)
		q.Set("flag", flag)
		q.Set("sign", sign.CalcSign(addr+flag))
		//2、创建url
		u, _ := url.Parse(centerAddr + r.RequestURI)
		//3、生成完整url
		u.RawQuery = q.Encode()
		errCode = email.SendMail("Verify Email", addr, u.String(), language)
	}
}
