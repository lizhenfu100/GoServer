package logic

import (
	"common"
	"common/tool/email"
)

func Rpc_login_relay_email(req, ack *common.NetPack) {
	subject := req.ReadString()
	target := req.ReadString()
	body := req.ReadString()
	language := req.ReadString()

	email.SendMail2(subject, target, body, language)
}
