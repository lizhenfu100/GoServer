package player

import (
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	g_player_cache  sync.Map //make(map[uint32]*TPlayer, 5000)
	g_account_cache sync.Map //make(map[uint32]*TPlayer, 5000)
)

func InitDB() {
	//只载入一个月内登录过的
	var list []TPlayer
	dbmgo.FindAll("Player", bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - 30*24*3600}}, &list)
	for i := 0; i < len(list); i++ {
		AddPlayerCache(&list[i])
	}
	println("load active player form db: ", len(list))

	InitSvrMailDB()
}

//! 多线程架构，玩家内存，只能他自己直接修改，别人须转给他后间接改(异步)
//! 其它玩家拿到指针，只允许 readonly
func FindPlayerInCache(pid uint32) *TPlayer {
	if v, ok := g_player_cache.Load(pid); ok {
		return v.(*TPlayer)
	}
	return nil
}
func FindWithDB_PlayerId(pid uint32) *TPlayer {
	if player := FindPlayerInCache(pid); player != nil {
		return player
	} else {
		if player := LoadPlayerFromDB("_id", pid); player != nil {
			AddPlayerCache(player)
			return player
		}
	}
	return nil
}
func FindWithDB_AccountId(aid uint32) *TPlayer {
	if v, ok := g_account_cache.Load(aid); ok {
		return v.(*TPlayer)
	} else {
		if player := LoadPlayerFromDB("accountid", aid); player != nil {
			AddPlayerCache(player)
			return player
		}
	}
	return nil
}
func AddNewPlayer(accountId uint32, name string) *TPlayer {
	playerId := dbmgo.GetNextIncId("PlayerId")
	player := _NewPlayerInDB(accountId, playerId, name)
	if player != nil {
		AddPlayerCache(player)
	}
	return player
}

// -------------------------------------
//! 辅助函数
func AddPlayerCache(player *TPlayer) {
	g_player_cache.Store(player.PlayerID, player)
	g_account_cache.Store(player.AccountID, player)
}
func DelPlayerCache(player *TPlayer) {
	g_player_cache.Delete(player.PlayerID)
	g_account_cache.Delete(player.AccountID)
}

// 多线程环境，做全服遍历，找死(╰_╯)#
// func ForEachOnlinePlayer(fun func(player *TPlayer)) {
// 	for _, v := range g_player_cache {
// 		if v.isOnlie {
// 			fun(v)
// 		}
// 	}
// }
