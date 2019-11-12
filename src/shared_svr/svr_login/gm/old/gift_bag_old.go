package old

import (
	"common"
	"common/std"
	"common/std/hash"
	"dbmgo"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"math"
	"strconv"
	"time"
)

const (
	kDBGift = "gift"
)

type TGiftBag struct {
	Key    string `bson:"_id"`
	Pf_id  string //平台名，相应平台玩家才能领，空则所有人可领
	Begin  int64  //在Begin-End之间才能领取此份奖励
	End    int64
	Json   string
	Pids   std.Strings //领过的人，pid、平台uuid
	MaxNum int         //限制领取总次数，默认无限
	Time   int64
}

func Rpc_login_get_gift(req, ack *common.NetPack) {
	code := req.ReadString()
	uuid := req.ReadString()
	pf_id := req.ReadString()

	key, timenow := GetDBKey(code), time.Now().Unix()
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
func getGift(key string) *TGiftBag {
	p := &TGiftBag{}
	if ok, _ := dbmgo.Find(kDBGift, "_id", key, p); ok {
		return p
	}
	return nil
}
