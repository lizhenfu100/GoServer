package logic

import (
	"common"
	"common/std/sign"
	"common/tool/email"
	"dbmgo"
	"fmt"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"netConfig/meta"
	"strconv"
	"sync"
	"time"
)

// 邮件解除设备的绑定关系
func Http_unbind_mac(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//不减少TSaveData.MacCnt，让解绑也时间受限制（无法准确区分是否自己操作）
	mac := q.Get("mac")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

	//! 创建回复
	ack := "Error: unknown"
	defer func() {
		w.Write(common.S2B(ack))
	}()

	if sign.CalcSign(mac+flag) != q.Get("sign") {
		ack = "Error: sign failed"
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack = "Error: url expire"
	} else if !dbmgo.RemoveOneSync(kDBMac, bson.M{"_id": mac}) {
		ack = "Error: DB Remove failed"
	} else {
		ack = "Unbind ok"
	}
}
func Rpc_save_unbind_mac_by_email(req, ack *common.NetPack) {
	mac := req.ReadString()
	emailAddr := req.ReadString()
	language := req.ReadString()

	if emailAddr == "None" { //无账号的渠道玩家，直接解绑
		dbmgo.RemoveOneSync(kDBMac, bson.M{"_id": mac})
	} else {
		//1、创建url
		httpAddr := fmt.Sprintf("http://%s:%d/unbind_mac",
			meta.G_Local.OutIP, meta.G_Local.Port())
		u, _ := url.Parse(httpAddr)
		q := u.Query()
		//2、写入参数
		q.Set("mac", mac)
		flag := strconv.FormatInt(time.Now().Unix(), 10)
		q.Set("flag", flag)
		q.Set("sign", sign.CalcSign(mac+flag))
		//3、生成完整url
		u.RawQuery = q.Encode()
		email.SendMail("Unbind Device", emailAddr, u.String(), language)
	}
}

var g_unbindTime sync.Map //<TSaveData.Key, int64>

func Rpc_save_unbind_mac(req, ack *common.NetPack) {
	mac, ptr := req.ReadString(), &MacInfo{}

	if ok, _ := dbmgo.Find(kDBMac, "_id", mac, ptr); ok {
		timeNow := time.Now().Unix()
		if timeOld, ok := g_unbindTime.Load(ptr.Key); ok {
			if timeNow-timeOld.(int64) < 3*24*3600 { //解绑，三天一次
				ack.WriteUInt16(err.Operate_too_often)
				return
			}
		}
		g_unbindTime.Store(ptr.Key, timeNow)
		dbmgo.RemoveOneSync(kDBMac, bson.M{"_id": mac})
	}
	ack.WriteUInt16(err.Success)
}
