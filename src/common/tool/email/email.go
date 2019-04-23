/***********************************************************************
* @ 使用第三方服务转发email
* @ brief
	、应交由固定ip节点来转发，如center/login，否则容易被第三方当成异地登录，临时封禁~囧
		· 其它节点须转到login发

	、国际服gmail，国内qq，天朝用gmail转发超时严重~真是蛋疼啊

	gmail：smtp.gmail.com, 465
		1、用户设置 - 转发和POP/IMAP - POP下载 - 对从现在起收到的邮件启用POP
		2、还须开启“安全性较低的应用的访问权限”
	qq：smtp.qq.com, 465
		1、用户设置 - 帐户 - POP3/IMAP/SMTP/Exchange/CardDAV/CalDAV服务 - POP3/SMTP服务
		2、取得验证码，替代gomail中的kPasswd
	aliyun：smtpdm.aliyun.com, 465
		1、

* @ Notice
	、不能用平常的邮箱，易检查到异地登录（自己用、后台也在用）
		· 分批次，每个大区各用一批
		· 该邮箱固定在某个ip下使用，防被封

* @ author zhoumf
* @ date 2018-11-15
***********************************************************************/
package email

import (
	"bytes"
	"common"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"gopkg.in/gomail"
	"math/rand"
	"netConfig"
	"text/template"
)

var g_list []*gomail.Dialer

func SendMail(subject, target, body, language string) {
	netConfig.CallRpcLogin(enum.Rpc_login_relay_email, func(buf *common.NetPack) {
		buf.WriteString(subject)
		buf.WriteString(target)
		buf.WriteString(body)
		buf.WriteString(language)
	}, nil)
}
func SendMail2(subject, target, body, language string) { //仅center/login调，其它节点转至login发送
	if g_list == nil {
		g_list = make([]*gomail.Dialer, len(conf.SvrCsv.EmailUser))
	}
	idx := rand.Intn(len(g_list))
	dialer := g_list[idx]
	if dialer == nil || dialer.Password != conf.SvrCsv.EmailPasswd[idx] {
		dialer = gomail.NewDialer(
			conf.SvrCsv.EmailHost,
			conf.SvrCsv.EmailPort,
			conf.SvrCsv.EmailUser[idx],
			conf.SvrCsv.EmailPasswd[idx])
		g_list[idx] = dialer
	}
	packBody(&subject, &body, language) //嵌入模板，并本地化

	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", dialer.Username, "ChillyRoom")
	msg.SetHeader("To", target)
	//msg.SetHeader("Cc" /*抄送*/, "xxxx@foxmail.com")
	//msg.SetHeader("Bcc" /*暗送*/, "xxxx@gmail.com")
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	//msg.Attach("我是附件")

	if e := dialer.DialAndSend(msg); e != nil {
		gamelog.Error("SendMail: %s", e.Error())
	}
}

func packBody(subject, body *string, language string) {
	content := translate(*subject, language)
	if content == "" {
		language = conf.SvrCsv.EmailLanguage
		content = translate(*subject, language)
	}
	if content != "" {
		if t, e := template.New(*subject).Parse(content); e == nil {
			var bf bytes.Buffer
			t.Execute(&bf, body)
			*body = bf.String()
		}
	}
	if content := translate(*subject+" Ex", language); content != "" {
		*subject = content
	}
	return
}
