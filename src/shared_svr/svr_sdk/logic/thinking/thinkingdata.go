package thinking

import (
	"gamelog"
	"github.com/ThinkingDataAnalytics/go-sdk/thinkingdata"
	"reflect"
	"shared_svr/svr_sdk/conf"
	"shared_svr/svr_sdk/msg"
	"time"
)

const (
	event_name = "order_finish"
	kURL       = "https://dc.chillyroom.com"
)

var _app = map[string]thinkingdata.TDAnalytics{}

func InitApp() {
	for _, v := range conf.ThinkingApp() {
		if p, e := thinkingdata.NewDebugConsumer(kURL, v.AppId); e != nil {
			gamelog.Error("%s: %s", v.Game, e.Error())
		} else {
			_app[v.Game] = thinkingdata.New(p)
		}
	}
}
func Track(p *msg.TOrderInfo) {
	data := map[string]interface{}{"#time": time.Now().UTC()}
	typ, val := reflect.TypeOf(*p), reflect.ValueOf(p).Elem()
	for i := 0; i < typ.NumField(); i++ {
		switch field := typ.Field(i).Name; field {
		case //不需要的字段
			"App_id", "Server_id", "Role_id",
			"Status", "Can_send",
			"Imsi", "Imei", "Ip", "Net",
			"Time", "Extra":
			continue
		default:
			data[field] = val.Field(i).Interface()
		}
	}
	if t, ok := _app[p.Game_id]; !ok {
		gamelog.Error("None game: %s", p.Game_id)
	} else if e := t.Track(p.Account, p.Mac, event_name, data); e != nil {
		gamelog.Error(e.Error())
	}
}
