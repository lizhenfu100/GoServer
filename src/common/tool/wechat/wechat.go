/***********************************************************************
* @ 微信报警
* @ brief
    1、申请开通企业微信
	2、“我的企业” -> 【企业ID】
    3、“我的企业” -> 自建应用【AgentId、Secret】
    4、“我的企业” -> “微工作台” -> 扫二维码关注，即可在微信中接收企业通知

* @ author zhoumf
* @ date 2019-3-5
***********************************************************************/
package wechat

import (
	"common"
	"common/assert"
	"encoding/json"
	"errors"
	"fmt"
	"gamelog"
	"netConfig/meta"
	"nets/http"
)

var (
	g_corpId  string
	g_secret  string
	g_agentId int    //企业微信中的应用id
	g_touser  string //消息接收者，多个用‘|’分隔，可指定为@all
	g_token   string //有过期时间
)

func Init(corpId, secret, touser string, agentId int) {
	g_corpId = corpId
	g_secret = secret
	g_touser = touser
	g_agentId = agentId

	if e := updateToken(corpId, secret); e != nil {
		fmt.Println("Wechat token err: ", e.Error())
	}
}
func SendMsg(text string) {
	if assert.IsDebug {
		return
	}
	buf, _ := json.Marshal(&msgWechat{
		Agentid: g_agentId,
		Touser:  g_touser,
		Msgtype: "text",
		Text:    map[string]string{"content": format(text)},
	})
	if e := sendMsg(buf); e != nil {
		if e = updateToken(g_corpId, g_secret); e != nil {
			gamelog.Error("Wechat token: ", e.Error())
		} else if e = sendMsg(buf); e != nil {
			gamelog.Error("Wechat send: %s", e.Error())
		}
	}
}
func format(text string) string {
	if meta.G_Local == nil {
		return "test\n--------------------------\n" + text
	}
	return fmt.Sprintf("%s(%d) %s %s",
		meta.G_Local.Module,
		meta.G_Local.SvrID,
		meta.G_Local.SvrName,
		meta.G_Local.OutIP) +
		"\n--------------------------\n" + text
}

// ------------------------------------------------------------
const (
	kUrlSend     = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token="
	kUrlGetToken = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid="
)

//定义一个简单的文本消息格式
type msgWechat struct {
	Agentid int               `json:"agentid"`
	Touser  string            `json:"touser"`  //消息接收者，多个用‘|’分隔，指定为@all，则向全部成员发送
	Toparty string            `json:"toparty"` //部门，多个用‘|’分隔，最多支持100个，当touser为@all时忽略本参数
	Totag   string            `json:"totag"`   //标签，多个用‘|’分隔，最多支持100个，当touser为@all时忽略本参数
	Safe    int               `json:"safe"`    //是否保密消息
	Msgtype string            `json:"msgtype"`
	Text    map[string]string `json:"text"`
}
type token struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"` //token有效秒数
}
type errMsg struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func updateToken(corpId, secret string) error {
	if buf := http.Client.Get(kUrlGetToken + corpId + "&corpsecret=" + secret); buf == nil {
		return http.ErrGet
	} else {
		var val token
		json.Unmarshal(buf, &val)
		if g_token = val.Access_token; g_token == "" {
			return errors.New(common.B2S(buf))
		}
	}
	return nil
}
func sendMsg(b []byte) error {
	if buf := http.Client.Post(kUrlSend+g_token, "application/json", b); buf == nil {
		return http.ErrPost
	} else {
		var msg errMsg
		json.Unmarshal(buf, &msg)
		if msg.Errcode != 0 && msg.Errmsg != "ok" {
			return errors.New(common.B2S(buf))
		}
	}
	return nil
}
