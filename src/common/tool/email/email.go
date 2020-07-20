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
	"common/assert"
	"common/format"
	"common/timer"
	"conf"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"gopkg.in/gomail"
	"math/rand"
	"netConfig"
	"text/template"
)

var G_Switch = true

func SendByLogin(subject, addr, body, language string) (errcode uint16) {
	if p, ok := netConfig.GetRpcRand(conf.GameName); ok {
		p.CallRpc(enum.Rpc_login_send_email, func(buf *common.NetPack) {
			buf.WriteString(subject)
			buf.WriteString(addr)
			buf.WriteString(body)
			buf.WriteString(language)
		}, func(recvBuf *common.NetPack) {
			errcode = recvBuf.ReadUInt16()
		})
	}
	return
}

// 仅center/login调，其它节点由login转发
func SendMail(subject, addr, body, language string) (errcode uint16) {
	if !format.CheckBindValue("email", addr) {
		return err.Invalid
		//TODO:return err.Email_format_err 2019春节版本后弃用旧错误码
	}
	if Invalid(addr) {
		return err.Invalid
		//TODO:return err.Is_forbidden
	}
	if !assert.IsDebug && !g_freq.Check(subject+addr) { //同内容的，限制发送频率
		return err.Success
		//TODO:return err.Email_try_send_please_check
	}
	if !G_Switch {
		return err.Closed_by_gm
	}
	packBody(&subject, &body, language) //嵌入模板，并本地化

	msg, dialer := gomail.NewMessage(), dialer()
	msg.SetAddressHeader("From", dialer.Username, "ChillyRoom")
	msg.SetHeader("To", addr)
	//msg.SetHeader("Cc" /*抄送*/, "xxxx@foxmail.com")
	//msg.SetHeader("Bcc" /*暗送*/, "xxxx@gmail.com")
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	//msg.Attach("我是附件")

	if e := dialer.DialAndSend(msg); e != nil {
		gamelog.Warn("%s:%s: %s", addr, subject, e.Error())
		return err.Invalid
		//TODO:return err.Email_unreachable
	}
	return err.Success
}

// ------------------------------------------------------------
var (
	g_list []*gomail.Dialer
	g_freq = timer.NewFreq(1, 240)
)

func dialer() *gomail.Dialer {
	csv := conf.SvrCsv()
	if g_list == nil {
		g_list = make([]*gomail.Dialer, len(csv.EmailUser))
	}
	idx := rand.Intn(len(g_list))
	ret := g_list[idx]
	if ret == nil || ret.Password != csv.EmailPasswd[idx] {
		ret = gomail.NewDialer(
			csv.EmailHost,
			csv.EmailPort,
			csv.EmailUser[idx],
			csv.EmailPasswd[idx])
		g_list[idx] = ret
	}
	return ret
}
func packBody(subject, body *string, language string) {
	if body2, ok := Translate(*subject, language); ok {
		if t, e := template.New(*subject).Parse(body2); e == nil {
			var bf bytes.Buffer
			t.Execute(&bf, body)
			*body = bf.String()
		}
	}
	if title, ok := Translate(*subject+" Ex", language); ok {
		*subject = title
	}
	return
}
