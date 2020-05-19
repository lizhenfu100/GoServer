/***********************************************************************
* @ 礼包码
* @ brief

* @ 接口文档
	· Rpc_login_get_gift
	· 上行参数
		· string key        礼包码key
		· uint32 pid        玩家playerId，可hash(uuid)代替
		· string pf_id      平台名，有些礼包仅固定平台领取
		· string version    客户端版本号，小于礼包版本，无法领
	· 下行参数
		· uint16 errCode
		· string json       客户端自行解析

* @ author zhoumf
* @ date 2018-12-12
***********************************************************************/
package gift_bag

import (
	"common"
	"common/copy"
	"common/std"
	"common/std/hash"
	"conf"
	"dbmgo"
	"encoding/json"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"math"
	"net/http"
	"strconv"
	"time"
)

const (
	KDBGift       = "gift"
	KDBGiftCode   = "giftcode"
	KDBGiftPlayer = "giftuid"
)

type TGiftBag struct {
	Key     string `bson:"_id"`
	Pf_id   string //平台名，相应平台玩家才能领，空则所有人可领
	Begin   int64  //在Begin-End之间才能领取此份奖励
	End     int64
	Json    string
	Version string
	Time    int64 `json:"-"`
}
type TGiftCode struct {
	Code string `bson:"_id"` //已被领的礼包码
	Uuid string
	Time int64
}
type TGiftPlayer struct { //玩家领取过的礼包
	Uuid string `bson:"_id"`
	Keys std.Strings
}

// ------------------------------------------------------------
func Rpc_login_gift_get(req, ack *common.NetPack, _ common.Conn) {
	code := req.ReadString() //礼包码
	uuid := req.ReadString()
	pf_id := req.ReadString()
	version := req.ReadString()
	key, timenow := GetDBKey(code), time.Now().Unix() //通过礼包码获得key
	if p := getGift(key); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else if common.CompareVersion(version, p.Version) < 0 {
		ack.WriteUInt16(err.Version_not_match)
	} else if p.Pf_id != "" && p.Pf_id != pf_id {
		ack.WriteUInt16(err.Pf_id_not_match) //非此平台玩家
	} else if !common.InTime(timenow, p.Begin, p.End) {
		ack.WriteUInt16(err.Not_in_the_time_zone) //时间不对，无法领取
	} else if ok, _ := dbmgo.Find(KDBGiftCode, "_id", code, &TGiftCode{}); ok { //礼包码数据库找到代表被领取
		ack.WriteUInt16(err.Already_get_it) //已领过
	} else if _playerGot(key, uuid) {
		ack.WriteUInt16(err.Already_get_it) //已领过
	} else {
		dbmgo.Insert(KDBGiftCode, &TGiftCode{code, uuid, timenow})
		ack.WriteUInt16(err.Success)
		ack.WriteString(p.Json)
	}
}
func _playerGot(key, uuid string) bool { //一类礼包只能领一次
	p := &TGiftPlayer{Uuid: uuid}
	if ok, _ := dbmgo.Find(KDBGiftPlayer, "_id", uuid, p); !ok {
		p.Keys = append(p.Keys, key)
		dbmgo.Insert(KDBGiftPlayer, p)
		return false
	} else if p.Keys.Index(key) < 0 {
		p.Keys = append(p.Keys, key)
		dbmgo.UpdateId(KDBGiftPlayer, uuid, bson.M{"$push": bson.M{"keys": key}})
		return false
	} else {
		return true
	}
}

// ------------------------------------------------------------
// -- 追加前后缀，作为礼包码
const (
	kPrefixLen = 4 //前缀：随机字符串
	kSuffixLen = 3 //后缀：数字校验码（值可能个位数，0开头）
	kSuffixFmt = "%03d"
)

func GetDBKey(code string) string {
	if kLen := len(code); kLen > kPrefixLen+kSuffixLen {
		suffix := code[kLen-kSuffixLen:] //后缀，是校验码
		if v, e := strconv.Atoi(suffix); e == nil {
			kMod := uint32(math.Pow10(kSuffixLen))
			if hash.StrHash(code[:kLen-kSuffixLen])%kMod == uint32(v) {
				return code[kPrefixLen : kLen-kSuffixLen]
			}
		}
	}
	return code
}

// ------------------------------------------------------------
func Http_gift_bag_add(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	p := &TGiftBag{Time: time.Now().Unix()}
	copy.CopyForm(p, q)

	if p.Key == "" {
		p.Key = strconv.Itoa(int(dbmgo.GetNextIncId("GiftId")))
	}
	if !dbmgo.InsertSync(KDBGift, p) {
		w.Write([]byte("gift repeat"))
	} else {
		gamelog.Info("Http_gift_bag_add: %v", q)
		w.Write([]byte("add ok: \n\n"))
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	}
}
func Http_gift_bag_set(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
	} else if p := getGift(q.Get("key")); p == nil {
		w.Write([]byte("none gift"))
	} else {
		gamelog.Info("Http_gift_bag_set: %v\n%v", q, p)
		copy.CopyForm(p, q)
		dbmgo.UpdateId(KDBGift, p.Key, p)
		w.Write([]byte("set ok: \n\n"))
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	}
}
func Http_gift_bag_view(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if p := getGift(q.Get("key")); p == nil {
		w.Write([]byte("none gift"))
	} else {
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	}
}
func Http_gift_bag_del(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
	} else if p := getGift(q.Get("key")); p == nil {
		w.Write([]byte("none gift"))
	} else {
		gamelog.Info("Http_gift_bag_del: %v", p)
		dbmgo.Remove(KDBGift, bson.M{"_id": p.Key})
		w.Write([]byte("del ok: \n\n"))
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	}
}
func Http_gift_bag_clear(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	timestamp, _ := strconv.ParseInt(q.Get("time"), 10, 64)

	if timestamp > time.Now().Unix()-3600*24*30 {
		w.Write([]byte("timestamp error"))
	} else {
		dbmgo.RemoveAll(KDBGift, bson.M{"time": bson.M{"$lt": timestamp}})
		dbmgo.RemoveAll(KDBGiftCode, bson.M{"time": bson.M{"$lt": timestamp}})
		w.Write([]byte("ok"))
	}
}

// ------------------------------------------------------------
func getGift(key string) *TGiftBag {
	p := &TGiftBag{}
	if ok, _ := dbmgo.Find(KDBGift, "_id", key, p); ok {
		return p
	}
	return nil
}
