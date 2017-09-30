package player

import (
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

const (
	Login_Active_Time = 30 * 24 * 3600 //一个月内登录过的，算活跃玩家
)

var (
	g_player_mutex  sync.RWMutex
	g_player_cache  = make(map[uint32]*TPlayer, 5000)
	g_account_cache = make(map[uint32]*TPlayer, 5000)
)

func LoadActivePlayerFormDB() {
	//只载入活跃玩家
	var playerLst []TPlayer
	dbmgo.FindAll("Player", bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - Login_Active_Time}}, &playerLst)
	for i := 0; i < len(playerLst); i++ {
		AddPlayerCache(&playerLst[i])
	}
	println("load active player form db: ", len(playerLst))
}

//! 多线程架构，玩家内存，只能他自己直接修改，别人须转给他后间接改(异步)
//! 其它玩家拿到指针，只允许 readonly
func FindPlayerInCache(id uint32) *TPlayer {
	g_player_mutex.RLock()
	ret := g_player_cache[id]
	g_player_mutex.RUnlock()
	return ret
}
func FindWithDB_PlayerId(id uint32) *TPlayer {
	if player := FindPlayerInCache(id); player != nil {
		return player
	} else {
		if player := LoadPlayerFromDB("_id", id); player != nil {
			AddPlayerCache(player)
			return player
		}
	}
	return nil
}
func FindWithDB_AccountId(id uint32) *TPlayer {
	g_player_mutex.RLock()
	player := g_account_cache[id]
	g_player_mutex.RUnlock()

	if player != nil {
		return player
	} else {
		if player = LoadPlayerFromDB("accountid", id); player != nil {
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
	g_player_mutex.Lock()
	g_player_cache[player.PlayerID] = player
	g_account_cache[player.AccountID] = player
	g_player_mutex.Unlock()
}
func DelPlayerCache(player *TPlayer) {
	g_player_mutex.Lock()
	delete(g_player_cache, player.PlayerID)
	delete(g_account_cache, player.AccountID)
	g_player_mutex.Unlock()
}

// 多线程环境，做全服遍历，找死(╰_╯)#
// func ForEachOnlinePlayer(fun func(player *TPlayer)) {
// 	for _, v := range g_player_cache {
// 		if v.isOnlie {
// 			fun(v)
// 		}
// 	}
// }
