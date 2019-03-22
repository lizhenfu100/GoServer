/***********************************************************************
* @ 给单个玩家发奖励
* @ brief
	、用于购买未到账等异常情况

* @ 单个玩家补偿
	、就是封邮件，只是前端可能没专门界面显示罢了
	、无邮件系统的游戏，进游戏时询问后台，是否有该玩家的邮件

* @ author zhoumf
* @ date 2019-1-28
***********************************************************************/
package gm

import (
	"common"
	"conf"
	"dbmgo"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"nets/tcp"
	"time"
)

const (
	kDBPlayerAward = "PlayerAward"
)

type PlayerAward struct {
	ID      uint32 `bson:"_id"`
	ToUid   string //给谁的，可以是账号名等
	Json    string //奖励内容
	Time    int64
	IsValid bool
}

// ------------------------------------------------------------
func Rpc_game_show_player_award(req, ack *common.NetPack, conn *tcp.TCPConn) {
	uid := req.ReadString()

	var list []*PlayerAward
	g_awards.Range(func(k, v interface{}) bool {
		p := v.(*PlayerAward)
		if uid == p.ToUid && p.IsValid == true {
			list = append(list, p)
		}
		return true
	})

	ack.WriteByte(byte(len(list)))
	for _, v := range list {
		ack.WriteUInt32(v.ID)
		ack.WriteString(v.Json)
	}
}
func Rpc_game_get_player_award_ok(req, ack *common.NetPack, conn *tcp.TCPConn) {
	id := req.ReadUInt32()
	if p := getAward(id); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else if p.IsValid == false {
		ack.WriteUInt16(err.Already_get_it)
	} else {
		dbmgo.UpdateId(kDBPlayerAward, id, bson.M{"$set": bson.M{"isvalid": false}})
		p.IsValid = false
		ack.WriteUInt16(err.Success)
	}
}

// ------------------------------------------------------------
var g_awards sync.Map //<id, *PlayerAward>

func InitAwardDB() {
	//删除超过30天的
	dbmgo.RemoveAllSync(kDBPlayerAward, bson.M{"time": bson.M{"$lt": time.Now().Unix() - 30*24*3600}})
	//载入所有未完成的
	var list []PlayerAward
	dbmgo.FindAll(kDBPlayerAward, bson.M{"isvalid": true}, &list)
	for i := 0; i < len(list); i++ {
		g_awards.Store(list[i].ID, &list[i])
	}
	println("load active PlayerAward form db: ", len(list))
}
func Rpc_game_gm_add_award(req, ack *common.NetPack, conn *tcp.TCPConn) {
	passwd := req.ReadString()
	touid := req.ReadString()
	Json := req.ReadString()

	if passwd != conf.GM_Passwd {
		ack.WriteUInt16(err.Passwd_err)
		return
	}
	p := &PlayerAward{
		dbmgo.GetNextIncId("AwardId"),
		touid,
		Json,
		time.Now().Unix(),
		true,
	}
	if dbmgo.InsertSync(kDBPlayerAward, p) {
		g_awards.Store(p.ID, p)
		ack.WriteUInt16(err.Success)
		gamelog.Info("gm_add_award: %s %s", touid, Json)
	} else {
		ack.WriteUInt16(err.Data_repeat)
	}

	//删除内存滞留的无效数据
	g_awards.Range(func(k, v interface{}) bool {
		if v.(*PlayerAward).IsValid == false {
			g_awards.Delete(k)
		}
		return true
	})
}
func Rpc_game_gm_set_award(req, ack *common.NetPack, conn *tcp.TCPConn) {
	passwd := req.ReadString()
	id := req.ReadUInt32()
	touid := req.ReadString()
	Json := req.ReadString()

	if passwd != conf.GM_Passwd {
		ack.WriteUInt16(err.Passwd_err)
		return
	}
	if p := getAward(id); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		gamelog.Info("gm_set_award: %d %s %s\n%v", id, touid, Json, p)
		p.ToUid = touid
		p.Json = Json
		dbmgo.UpdateId(kDBPlayerAward, p.ID, p)
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_game_gm_del_award(req, ack *common.NetPack, conn *tcp.TCPConn) {
	passwd := req.ReadString()
	id := req.ReadUInt32()

	if passwd != conf.GM_Passwd {
		ack.WriteUInt16(err.Passwd_err)
		return
	}
	if p := getAward(id); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		g_awards.Delete(p.ID)
		dbmgo.Remove(kDBPlayerAward, bson.M{"_id": p.ID})
		ack.WriteUInt16(err.Success)
		gamelog.Info("gm_del_award: %v", p)
	}
}

// ------------------------------------------------------------
func getAward(id uint32) *PlayerAward {
	if v, ok := g_awards.Load(id); ok {
		return v.(*PlayerAward)
	}
	return nil
}
