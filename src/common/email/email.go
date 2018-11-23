/***********************************************************************
* @ 使用第三方服务转发Email
* @ brief
	gmail：
		1、用户设置 - 转发和POP/IMAP - POP下载 - 对从现在起收到的邮件启用POP
		2、还须开启“安全性较低的应用的访问权限”
	qq：
		1、用户设置 - 帐户 - POP3/IMAP/SMTP/Exchange/CardDAV/CalDAV服务 - POP3/SMTP服务
		2、取得验证码，替代gomail中的kPasswd

* @ author zhoumf
* @ date 2018-11-15
***********************************************************************/
package email

import (
	"gamelog"
	"gopkg.in/gomail"
)

const (
	//kUser        = "734688714@qq.com"
	//kPasswd      = "ezblhqfudwfabead"
	//kHost, kPort = "smtp.qq.com", 465
	kUser        = "3workman@gmail.com"
	kHost, kPort = "smtp.gmail.com", 465
)

var (
	g_msg    = gomail.NewMessage()
	g_dialer = gomail.NewDialer(kHost, kPort, kUser, "zmf890104")
)

func SendMail(subject, target, body string) {
	g_msg.Reset()
	msg := g_msg

	msg.SetAddressHeader("From", kUser, "ChillyRoom")
	msg.SetHeader("To", target)
	//msg.SetHeader("Cc" /*抄送*/, "xxxx@foxmail.com")
	//msg.SetHeader("Bcc" /*暗送*/, "xxxx@gmail.com")
	msg.SetHeader("Subject", subject)

	msg.SetBody("text/html", body)

	//msg.Attach("我是附件")

	if err := g_dialer.DialAndSend(msg); err != nil {
		gamelog.Error("SendMail err: %s", err.Error())
	}
}
