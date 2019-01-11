package player

import (
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"svr_game/logic/fullsvr"
	"sync"
	"time"
)

var (
	G_player_cache  sync.Map //<pid, *TPlayer>
	g_account_cache sync.Map //<aid, *TPlayer>
	g_online_cnt    int32
)

func InitDB() {
	//只载入一个月内登录过的
	var list1 []TPlayerBase
	dbmgo.FindAll(kDBPlayer, bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - 30*24*3600}}, &list1)
	list := make([]TPlayer, len(list1))
	for i := 0; i < len(list); i++ {
		ptr := &list[i]
		ptr.init()
		ptr.TPlayerBase = list1[i]
		for _, v := range ptr.modules {
			v.LoadFromDB(ptr)
		}
		AddCache(ptr)
	}
	println("load active player form db: ", len(list))

	InitSvrMailDB()
	InitSeasonDB()
	fullsvr.InitAwardDB()
}

//! 若多线程架构，玩家内存，只能他自己直接修改，别人须转给他后间接改(异步)
func FindPlayerId(pid uint32) *TPlayer {
	if v, ok := G_player_cache.Load(pid); ok {
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
	G_player_cache.Store(player.PlayerID, player)
	g_account_cache.Store(player.AccountID, player)
}
func DelCache(player *TPlayer) {
	G_player_cache.Delete(player.PlayerID)
	g_account_cache.Delete(player.AccountID)
}

// ------------------------------------------------------------
//! 访问玩家部分数据，包括离线的
func GetPlayerBase(pid uint32) *TPlayerBase {
	if player := FindPlayerId(pid); player != nil {
		return &player.TPlayerBase
	} else {
		ptr := new(TPlayerBase)
		if dbmgo.Find(kDBPlayer, "_id", pid, ptr) {
			return ptr
		}
		return nil
	}
}
