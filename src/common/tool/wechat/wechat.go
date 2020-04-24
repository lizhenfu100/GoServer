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
	"common/timer"
	"encoding/json"
	"fmt"
	"gamelog"
	"netConfig/meta"
	"nets/http"
)

var (
	_corpId  string //企业id
	_secret  string //应用secret
	_agentId int    //应用id
	g_freq   = timer.NewFreq(1, 60)
)

func Init(corpId, secret string, agentId int) {
	_corpId, _secret, _agentId = corpId, secret, agentId
	if e := updateToken(); e != nil {
		fmt.Println("Wechat init: ", e.Error())
	}
}
func SendMsg(text string) {
	if assert.IsDebug || !g_freq.Check(text) {
		return
	}
	buf, _ := json.Marshal(&msgWechat{
		Agentid: _agentId,
		Touser:  "@all",
		Msgtype: "text",
		Text:    map[string]string{"content": format(text)},
	})
	if e := sendMsg(buf); e != nil {
		if e = updateToken(); e != nil {
			gamelog.Error("%s %s", text, e.Error())
		} else if e = sendMsg(buf); e != nil {
			gamelog.Error("%s %s", text, e.Error())
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

var g_token token //有过期时间

func updateToken() error {
	if buf := http.Client.Get(kUrlGetToken + _corpId + "&corpsecret=" + _secret); buf == nil {
		return http.ErrGet
	} else {
		g_token.Access_token = ""
		json.Unmarshal(buf, &g_token)
		if g_token.Access_token == "" {
			return common.Err(common.B2S(buf))
		}
	}
	return nil
}
func sendMsg(b []byte) error {
	url := kUrlSend + g_token.Access_token
	if buf := http.Client.Post(url, "application/json", b); buf == nil {
		return http.ErrPost
	} else {
		var msg errMsg
		json.Unmarshal(buf, &msg)
		if msg.Errcode != 0 && msg.Errmsg != "ok" {
			return common.Err(common.B2S(buf))
		}
	}
	return nil
}
