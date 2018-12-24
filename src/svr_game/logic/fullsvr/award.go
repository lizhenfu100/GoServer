/***********************************************************************
* @ 全服奖励，礼包码
* @ brief
	、stAward.Key包含前缀，用于区分不同性质的奖励，如游戏全服奖励、渠道礼包...
		· fullsvr_	常规全服奖励，大家都能领
			· 客户端可能要展示之（从g_award筛选对应前缀的数据）
		、

* @ 单个玩家补偿
	、就是封邮件，只是前端可能没专门界面显示罢了
	、无邮件系统的游戏，进游戏时询问后台，是否有该玩家的邮件

* @ author zhoumf
* @ date 2018-12-12
***********************************************************************/
package fullsvr

import (
	"common"
	"common/file"
	"dbmgo"
	"fmt"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"sync"
	"tcp"
	"time"
)

const kDBAward = "Award"

var g_award sync.Map //<key, *stAward>

type stAward struct {
	Key    string `bson:"_id"`
	Begin  int64  //在Begin-End之间才能领取此份奖励
	End    int64
	Desc   string
	Json   string
	Pf_id  string   //平台名，相应平台玩家才能领，空则所有人可领
	MaxNum int      //限制领取总次数，默认无限
	Pids   []uint32 //平台的玩家uid若是string，hash为uint32
}

// ------------------------------------------------------------
func Rpc_game_fullsvr_get_award(req, ack *common.NetPack, conn *tcp.TCPConn) {
	key := req.ReadString()
	pid := req.ReadUInt32()
	pf_id := req.ReadString()

	if p := getAward(key); p == nil {
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
		dbmgo.UpdateId(kDBAward, key, bson.M{"$push": bson.M{"pids": pid}})
		ack.WriteUInt16(err.Success)
		ack.WriteString(p.Json)
	}
}

// ------------------------------------------------------------
// -- 各种前缀
func getNewKey() string { //全服奖励
	return fmt.Sprintf("fullsvr_%d", dbmgo.GetNextIncId("AwardId"))
}

// ------------------------------------------------------------
func InitAwardDB() {
	var list []stAward
	timenow := time.Now().Unix()
	dbmgo.FindAll(kDBAward, bson.M{"timeend": bson.M{"$gt": timenow}}, &list)
	for i := 0; i < len(list); i++ {
		g_award.Store(list[i].Key, &list[i])
	}
}
func Rpc_game_fullsvr_add_award(req, ack *common.NetPack, conn *tcp.TCPConn) {
	var data stAward
	ref := reflect.ValueOf(&data).Elem()

	cnt := req.ReadUInt8()
	for i := 0; i < int(cnt); i++ {
		field := req.ReadString()
		value := req.ReadString()
		file.SetField(ref.FieldByName(field), value)
	}

	if data.Key == "" {
		data.Key = getNewKey()
	} else if getAward(data.Key) != nil {
		ack.WriteUInt16(err.Record_repeat)
		return
	}
	if dbmgo.InsertSync(kDBAward, &data) {
		g_award.Store(data.Key, &data)
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Record_repeat)
	}

}
func Rpc_game_fullsvr_set_award(req, ack *common.NetPack, conn *tcp.TCPConn) {
	key := req.ReadString()

	if p := getAward(key); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		ref := reflect.ValueOf(p).Elem()

		cnt := req.ReadUInt8()
		for i := 0; i < int(cnt); i++ {
			field := req.ReadString()
			value := req.ReadString()
			file.SetField(ref.FieldByName(field), value)
			ack.WriteUInt16(err.Success)
		}
	}
}
func Rpc_game_fullsvr_del_award(req, ack *common.NetPack, conn *tcp.TCPConn) {
	key := req.ReadString()
	if p := getAward(key); p != nil {
		g_award.Delete(key)
		dbmgo.Remove(kDBAward, bson.M{"_id": key})
	}
}

// ------------------------------------------------------------
func getAward(key string) *stAward {
	if v, ok := g_award.Load(key); ok {
		return v.(*stAward)
	}
	return nil
}
func (self *stAward) havePid(pid uint32) bool {
	for _, v := range self.Pids {
		if pid == v {
			return true
		}
	}
	return false
}
