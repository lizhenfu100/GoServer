/***********************************************************************
* @ 礼包码
* @ brief

* @ author zhoumf
* @ date 2018-12-12
***********************************************************************/
package logic

import (
	"common"
	"common/copy"
	"conf"
	"dbmgo"
	"encoding/json"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	kDBGift = "gift"
)

var g_gifts sync.Map //<key, *TGiftBag>

type TGiftBag struct {
	Key    string `bson:"_id"` //可自定义
	Begin  int64  //在Begin-End之间才能领取此份奖励
	End    int64
	Desc   string
	Json   string
	Pf_id  string   //平台名，相应平台玩家才能领，空则所有人可领
	MaxNum int      //限制领取总次数，默认无限
	Pids   []uint32 //领过的人，平台的玩家uid可哈希为uint32
}

// ------------------------------------------------------------
func Rpc_login_get_gift(req, ack *common.NetPack) {
	key := req.ReadString()
	pid := req.ReadUInt32()
	pf_id := req.ReadString()

	if p := getGift(key); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else if p.MaxNum > 0 && len(p.Pids) >= p.MaxNum {
		ack.WriteUInt16(err.Not_found) //被领完了
	} else if p.Pf_id != "" && p.Pf_id != pf_id {
		ack.WriteUInt16(err.Pf_id_not_match) //非此平台玩家
	} else if timenow := time.Now().Unix(); timenow < p.Begin || timenow > p.End {
		ack.WriteUInt16(err.Not_in_the_time_zone) //时间不对，无法领取
	} else if p.havePid(pid) {
		ack.WriteUInt16(err.Already_get_it) //已领过
	} else {
		p.Pids = append(p.Pids, pid)
		dbmgo.UpdateId(kDBGift, key, bson.M{"$push": bson.M{"pids": pid}})
		ack.WriteUInt16(err.Success)
		ack.WriteString(p.Json)
	}
}

// ------------------------------------------------------------
func InitGiftDB() {
	var list []TGiftBag
	timenow := time.Now().Unix()
	dbmgo.FindAll(kDBGift, bson.M{"timeend": bson.M{"$gt": timenow}}, &list)
	for i := 0; i < len(list); i++ {
		g_gifts.Store(list[i].Key, &list[i])
	}
}
func Http_gift_bag_add(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Form.Get("passwd") != conf.GM_Passwd {
		w.Write(common.ToBytes("passwd error"))
		return
	}
	if p := getGift(r.Form.Get("key")); p != nil {
		w.Write(common.ToBytes("gift repeat"))
		return
	}

	p := &TGiftBag{}
	if copy.CopyForm(p, r.Form); p.Key == "" {
		p.Key = strconv.Itoa(int(dbmgo.GetNextIncId("GiftId")))
	}
	if dbmgo.InsertSync(kDBGift, p) {
		g_gifts.Store(p.Key, p)
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	} else {
		w.Write(common.ToBytes("gift repeat"))
	}
}
func Http_gift_bag_set(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Form.Get("passwd") != conf.GM_Passwd {
		w.Write(common.ToBytes("passwd error"))
		return
	}

	if p := getGift(r.Form.Get("key")); p == nil {
		w.Write(common.ToBytes("fail"))
	} else {
		copy.CopyForm(p, r.Form)
		dbmgo.UpdateId(kDBGift, p.Key, p)
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	}
}
func Http_gift_bag_del(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Form.Get("passwd") != conf.GM_Passwd {
		w.Write(common.ToBytes("passwd error"))
		return
	}

	if p := getGift(r.Form.Get("key")); p == nil {
		w.Write(common.ToBytes("not find"))
	} else {
		g_gifts.Delete(p.Key)
		dbmgo.Remove(kDBGift, bson.M{"_id": p.Key})
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	}
}

// ------------------------------------------------------------
func getGift(key string) *TGiftBag {
	if v, ok := g_gifts.Load(key); ok {
		return v.(*TGiftBag)
	}
	return nil
}
func (self *TGiftBag) havePid(pid uint32) bool {
	for _, v := range self.Pids {
		if pid == v {
			return true
		}
	}
	return false
}
