package sms

import (
	"common"
	"common/timer"
	"fmt"
	"gamelog"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"math/rand"
	"sync"
	"time"
)

const (
	kAccessKeyId  = "LTAI4FhegmMD7vSWQvHwCdfX"
	kAccessSecret = "7l8bl5PJEp1mnoGuIQwyiTGJR5L3B7"
)

type Code struct {
	V string
	T int64
}

var (
	g_freq    = timer.NewFreq(1, 120)
	g_strBase = []byte("0123456789")
	g_code    sync.Map
)

func SendCode(phone string) {
	if !g_freq.Check(phone) {
		return
	}
	client, _ := dysmsapi.NewClientWithAccessKey("cn-hangzhou", kAccessKeyId, kAccessSecret)
	req, code := dysmsapi.CreateSendSmsRequest(), MakeCode()
	req.Scheme = "https"
	req.PhoneNumbers = phone
	req.SignName = "凉屋游戏"
	req.TemplateCode = "SMS_184830407" //短信模板ID
	req.TemplateParam = fmt.Sprintf("{code:%s}", code.V)
	if ack, e := client.SendSms(req); e != nil {
		gamelog.Error("SendSms: %s", e.Error())
	} else if ack.Code != "OK" {
		gamelog.Error("", ack.Message)
	} else {
		g_code.Store(phone, code)
	}
}
func MakeCode() Code {
	s := make([]byte, 6)
	for i := 0; i < len(s); i++ {
		s[i] = g_strBase[rand.Intn(len(g_strBase))]
	}
	return Code{common.B2S(s), time.Now().Unix()}
}
func CheckCode(phone, code string) bool {
	if v, ok := g_code.Load(phone); ok {
		now, c := time.Now().Unix(), v.(Code)
		if c.V == code && now-c.T <= 300 {
			g_code.Delete(phone)
			return true
		}
	}
	return false
}
