/***********************************************************************
* @ 发消息给qq群
* @ brief

* @ author zhoumf
* @ date 2019-10-15
***********************************************************************/
package qq

import (
	"encoding/json"
	"gamelog"
	"net/url"
	"nets/http/http"
	"strconv"
)

const (
	//路由器里显示的ip，web返回的ip不能用
	kUrlSend = "http://192.168.1.50:5700/send_group_msg"
)

func SendMsg(qq int, text string) {
	u, _ := url.Parse(kUrlSend)
	q := u.Query()
	q.Set("group_id", strconv.Itoa(qq))
	q.Set("message", text)
	q.Set("auto_escape", "true")
	u.RawQuery = q.Encode()
	if buf := http.Client.Get(u.String()); buf != nil {
		var ack ackMsg
		if e := json.Unmarshal(buf, &ack); e != nil {
			gamelog.Error("QQ send: %s", e.Error())
		} else if ack.Status != "ok" {
			gamelog.Error("QQ send: %v", ack)
		}
	} else {
		gamelog.Error("QQ send fail: %s", text)
	}
}

type ackMsg struct {
	Status  string                 `json:"status"`
	Retcode int                    `json:"retcode"`
	Data    map[string]interface{} `json:"data"`
}
