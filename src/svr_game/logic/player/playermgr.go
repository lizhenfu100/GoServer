package player

import (
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"sync/atomic"
	"time"
)

var (
	g_player_cache  sync.Map //make(map[uint32]*TPlayer, 5000)
	g_account_cache sync.Map //make(map[uint32]*TPlayer, 5000)
	g_player_cnt    int32
)

func InitDB() {
	//只载入一个月内登录过的
	var list []TPlayer
	dbmgo.FindAll("Player", bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - 30*24*3600}}, &list)
	for i := 0; i < len(list); i++ {
		list[i].init()
		AddCache(&list[i])
	}
	println("load active player form db: ", len(list))

	InitSvrMailDB()
}

//! 若多线程架构，玩家内存，只能他自己直接修改，别人须转给他后间接改(异步)
func FindPlayerId(pid uint32) *TPlayer {
	if v, ok := g_player_cache.Load(pid); ok {
		return v.(*TPlayer)
	}
	return nil
}
func FindAccountId(aid uint32) *TPlayer {
	if v, ok := g_account_cache.Load(aid); ok {
		return v.(*TPlayer)
	}
	return nil
}

func FindWithDB_PlayerId(pid uint32) *TPlayer {
	if player := FindPlayerId(pid); player != nil {
		return player
	} else {
		return LoadPlayerFromDB("_id", pid)
	}
}
func FindWithDB_AccountId(aid uint32) *TPlayer {
	if player := FindAccountId(aid); player != nil {
		return player
	} else {
		return LoadPlayerFromDB("accountid", aid)
	}
}

// -------------------------------------
//! 辅助函数
func AddCache(player *TPlayer) {
	g_player_cache.Store(player.PlayerID, player)
	g_account_cache.Store(player.AccountID, player)
	atomic.AddInt32(&g_player_cnt, 1)
}
func DelCache(player *TPlayer) {
	g_player_cache.Delete(player.PlayerID)
	g_account_cache.Delete(player.AccountID)
	atomic.AddInt32(&g_player_cnt, -1)
}

// ------------------------------------------------------------
//! 访问玩家部分数据，包括离线的
func GetPlayerBaseData(pid uint32) *TPlayerBase {
	if player := FindPlayerId(pid); player != nil {
		return &player.TPlayerBase
	} else {
		ptr := new(TPlayerBase)
		if dbmgo.Find("Player", "_id", pid, ptr) {
			return ptr
		}
		return nil
	}
}
