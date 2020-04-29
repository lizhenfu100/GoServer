package sms

import (
	"bytes"
	"common"
	"common/timer"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gamelog"
	"math/rand"
	"net/url"
	"nets/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	g_freq  = timer.NewFreq(1, 120)
	g_code  sync.Map
	_keyId  string
	_secret string
)

type Code struct {
	V string
	T int64
}

func SendCode(phone string) {
	if !g_freq.Check(phone) {
		return
	}
	code := MakeCode()
	req, ack := MakeUrl(phone, code.V), smsAck{}
	if buf := http.Client.Get(req); buf == nil {
		gamelog.Error("SendSms: failed")
	} else if json.Unmarshal(buf, &ack); ack.Code != "OK" {
		gamelog.Error("%s : %s", ack.Code, ack.Message)
	} else {
		g_code.Store(phone, code)
	}
}
func MakeCode() Code {
	s := make([]byte, 6)
	for i := 0; i < len(s); i++ {
		s[i] = []byte("0123456789")[rand.Intn(10)]
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

// ------------------------------------------------------------
const (
	kUrlSend    = "https://dysmsapi.aliyuncs.com/?Signature="
	kTemplateId = "SMS_184830407"
	kName       = "凉屋游戏"
)

type smsAck struct {
	Code    string
	Message string
}

var _buf1, _buf2, _buf3 bytes.Buffer

func Init(key, secret string) {
	_keyId = key
	_secret = secret
	_buf1.WriteString("AccessKeyId=")
	_buf1.WriteString(encode(_keyId))
	_buf1.WriteString("&Action=SendSms")
	_buf1.WriteString("&Format=json")
	_buf1.WriteString("&PhoneNumbers=")
	//buf.WriteString(encode(phone))
	_buf2.WriteString("&RegionId=")
	_buf2.WriteString(encode("cn-hangzhou"))
	_buf2.WriteString("&SignName=")
	_buf2.WriteString(encode(kName))
	_buf2.WriteString("&SignatureMethod=")
	_buf2.WriteString(encode("HMAC-SHA1"))
	_buf2.WriteString("&SignatureNonce=")
	//buf.WriteString(encode(strconv.FormatUint(rand.Uint64(), 10)))
	_buf3.WriteString("&SignatureVersion=")
	_buf3.WriteString(encode("1.0"))
	_buf3.WriteString("&TemplateCode=")
	_buf3.WriteString(encode(kTemplateId))
	_buf3.WriteString("&TemplateParam=")
}
func MakeUrl(phone string, code string) string {
	var buf, sign, ret bytes.Buffer
	buf.Write(_buf1.Bytes())
	buf.WriteString(encode(phone))
	buf.Write(_buf2.Bytes())
	buf.WriteString(encode(strconv.FormatUint(rand.Uint64(), 10)))
	buf.Write(_buf3.Bytes())
	buf.WriteString(encode(fmt.Sprintf("{code:%s}", code)))
	buf.WriteString("&Timestamp=")
	buf.WriteString(encode(time.Now().UTC().Format("2006-01-02T15:04:05Z")))
	buf.WriteString("&Version=")
	buf.WriteString(encode("2017-05-25"))

	sign.WriteString("GET&")
	sign.WriteString(encode("/"))
	sign.WriteString("&")
	sign.WriteString(encode(buf.String()))
	mac := hmac.New(sha1.New, common.S2B(_secret+"&"))
	mac.Write(sign.Bytes())
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	ret.WriteString(kUrlSend)
	ret.WriteString(encode(signature))
	ret.WriteString("&")
	ret.Write(buf.Bytes())
	return ret.String()
}
func encode(v string) string {
	ret := url.QueryEscape(v)
	ret = strings.ReplaceAll(ret, "+", "%20")
	ret = strings.ReplaceAll(ret, "*", "%2A")
	ret = strings.ReplaceAll(ret, "%7E", "~")
	return ret
}
