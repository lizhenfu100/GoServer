/***********************************************************************
* @ 礼包码
* @ brief

* @ 接口文档
	· Rpc_login_get_gift
	· 上行参数
		· string key        礼包码key
		· uint32 pid        玩家playerId，可hash(uuid)代替
		· string pf_id      平台名，有些礼包仅固定平台领取
	· 下行参数
		· uint16 errCode
		· string json       客户端自行解析

* @ author zhoumf
* @ date 2018-12-12
***********************************************************************/
package gm

import (
	"common"
	"common/copy"
	"common/file"
	"common/std"
	"common/std/hash"
	"common/std/random"
	"conf"
	"dbmgo"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

const kDBGift = "gift"

type TGiftBag struct {
	Key    string `bson:"_id"` //可自定义
	Begin  int64  //在Begin-End之间才能领取此份奖励
	End    int64
	Desc   string
	Json   string
	Pf_id  string      //平台名，相应平台玩家才能领，空则所有人可领
	Pids   std.Strings //领过的人，pid、平台uuid
	MaxNum int         //限制领取总次数，默认无限
	Time   int64
}

// ------------------------------------------------------------
func Rpc_login_get_gift(req, ack *common.NetPack) {
	key := req.ReadString()
	uuid := req.ReadString()
	pf_id := req.ReadString()

	key = GetDBKey(key)
	timenow := time.Now().Unix()
	if p := getGift(key); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else if p.MaxNum > 0 && len(p.Pids) >= p.MaxNum {
		ack.WriteUInt16(err.Not_found) //被领完了
	} else if p.Pf_id != "" && p.Pf_id != pf_id {
		ack.WriteUInt16(err.Pf_id_not_match) //非此平台玩家
	} else if (p.Begin > 0 && timenow < p.Begin) || (p.End > 0 && timenow > p.End) {
		ack.WriteUInt16(err.Not_in_the_time_zone) //时间不对，无法领取
	} else if p.Pids.Index(uuid) >= 0 {
		ack.WriteUInt16(err.Already_get_it) //已领过
	} else {
		p.Pids = append(p.Pids, uuid)
		dbmgo.UpdateId(kDBGift, key, bson.M{"$push": bson.M{"pids": uuid}})
		ack.WriteUInt16(err.Success)
		ack.WriteString(p.Json)
	}
}

// ------------------------------------------------------------
// -- 特殊key，用于减少内容相同礼包的数量
const (
	kPrefixLen = 4 //4位前缀：随机字符串
	kSuffixLen = 4 //4位后缀：数字校验码（可能少于4位，0开头）
	kDBkeyLen  = 2
)

func GetDBKey(userkey string) (dbkey string) {
	if length := len(userkey); length > kPrefixLen+kDBkeyLen {
		str := userkey[:kPrefixLen+kDBkeyLen]
		suffix := userkey[kPrefixLen+kDBkeyLen:] //后缀，是校验码
		if v, e := strconv.Atoi(suffix); e == nil {
			kMod := uint32(math.Pow10(kSuffixLen))
			if hash.StrHash(str)%kMod == uint32(v) {
				return userkey[kPrefixLen : kPrefixLen+kDBkeyLen] //真正的dbkey
			}
		}
	}
	return userkey
}
func MakeUserKey(dbkey string) string {
	if len(dbkey) == kDBkeyLen {
		str, kMod := random.String(kPrefixLen)+dbkey, uint32(math.Pow10(kSuffixLen))
		sign := strconv.Itoa(int(hash.StrHash(str) % kMod))
		return str + sign
	} else {
		return dbkey
	}
}
func MakeUserKeyCsv(dbkey string, kAddNum int, csvDir, csvName string) {
	// 读已生成的key
	f, e := file.CreateFile(csvDir, csvName, os.O_APPEND|os.O_WRONLY)
	if e != nil {
		gamelog.Error("MakeUserKeyCsv: %s", e.Error())
	}
	records, e := file.ReadCsv(csvDir + csvName)
	if e != nil {
		gamelog.Error("MakeUserKeyCsv: %s", e.Error())
	}

	all := make(map[string]bool, len(records))
	for _, v := range records {
		all[v[0]] = true
	}
	news := make(map[string]bool, kAddNum)
	for {
		v := MakeUserKey(dbkey)
		if _, ok := all[v]; !ok {
			all[v] = true
			news[v] = true
		}
		fmt.Println("-------------------", len(all))
		if len(news) == kAddNum {
			break
		}
	}

	w, i := csv.NewWriter(f), 0
	for k := range news {
		if e := w.Write([]string{k}); e != nil {
			f.Close()
			return
		}
		if i++; i%1000 == 0 {
			w.Flush()
		}
	}
	w.Flush()
	f.Close()
	fmt.Println("------------------- csv end")
}

// ------------------------------------------------------------
func InitGiftDB() {
	//删除超过四个月的
	dbmgo.RemoveAllSync(kDBGift, bson.M{"time": bson.M{"$lt": time.Now().Unix() - 120*24*3600}})
}
func Http_gift_bag_add(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Form.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	if p := getGift(r.Form.Get("key")); p != nil {
		w.Write(common.S2B("gift repeat"))
		return
	}

	p := &TGiftBag{Time: time.Now().Unix()}
	if copy.CopyForm(p, r.Form); p.Key == "" {
		p.Key = strconv.Itoa(int(dbmgo.GetNextIncId("GiftId")))
	}
	if dbmgo.InsertSync(kDBGift, p) {
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
		gamelog.Info("Http_gift_bag_add: %v", r.Form)
	} else {
		w.Write(common.S2B("gift repeat"))
	}
}
func Http_gift_bag_set(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Form.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
	} else if p := getGift(r.Form.Get("key")); p == nil {
		w.Write(common.S2B("fail"))
	} else {
		gamelog.Info("Http_gift_bag_set: %v\n%v", r.Form, p)
		copy.CopyForm(p, r.Form)
		dbmgo.UpdateId(kDBGift, p.Key, p)
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
	}
}
func Http_gift_bag_del(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Form.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
	} else if p := getGift(r.Form.Get("key")); p == nil {
		w.Write(common.S2B("not find"))
	} else {
		dbmgo.Remove(kDBGift, bson.M{"_id": p.Key})
		ack, _ := json.MarshalIndent(p, "", "     ")
		w.Write(ack)
		gamelog.Info("Http_gift_bag_del: %v", p)
	}
}

// ------------------------------------------------------------
func getGift(key string) *TGiftBag {
	p := &TGiftBag{}
	if ok, _ := dbmgo.Find(kDBGift, "_id", key, p); ok {
		return p
	}
	return nil
}
