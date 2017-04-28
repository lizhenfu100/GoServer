package player

import (
	"dbmgo"
	"sync"
)

var (
	g_player_mutex  sync.RWMutex
	g_player_cache  = make(map[uint32]*TPlayer, 5000)
	g_account_cache = make(map[uint32]*TPlayer, 5000)
)

//! 多线程架构，玩家内存，只能他自己直接修改，别人须转给他后间接改(异步)
func _FindPlayerInCache(id uint32) *TPlayer {
	g_player_mutex.RLock()
	ret := g_player_cache[id]
	g_player_mutex.RUnlock()
	return ret
}
func FindWithDB_PlayerId(id uint32) *TPlayer {
	if player := _FindPlayerInCache(id); player != nil {
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
	if player := NewPlayerInDB(accountId, dbmgo.GetNextIncId("PlayerId"), name); player != nil {
		AddPlayerCache(player)
		return player
	}
	return nil
}
func AddPlayerCache(player *TPlayer) {
	g_player_mutex.Lock()
	g_player_cache[player.PlayerID] = player
	g_account_cache[player.AccountID] = player
	g_player_mutex.Unlock()
}
func DelPlayerCache(playerId uint32) {
	if player, ok := g_player_cache[playerId]; ok {
		g_player_mutex.Lock()
		delete(g_player_cache, player.PlayerID)
		delete(g_account_cache, player.AccountID)
		g_player_mutex.Unlock()
	}
}

// 多线程环境，做全服遍历，找死(╰_╯)#
// func ForEachOnlinePlayer(fun func(player *TPlayer)) {
// 	for _, v := range g_player_cache {
// 		if v.isOnlie {
// 			fun(v)
// 		}
// 	}
// }
