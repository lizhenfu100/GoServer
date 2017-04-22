package player

import (
	"dbmgo"
	"sync"
)

var (
	g_player_mutex  sync.Mutex
	g_player_cache  = make(map[uint32]*TPlayer, 5000)
	g_account_cache = make(map[uint32]*TPlayer, 5000)
	g_auto_playerId uint32
)

func FindPlayerInCache(id uint32) *TPlayer { return g_player_cache[id] }
func FindWithDB_PlayerId(id uint32) *TPlayer {
	if player, ok := g_player_cache[id]; ok {
		return player
	} else {
		player := new(TPlayer)
		if ok := player.LoadAllFromDB("_id", id); ok {
			AddPlayerCache(player)
			return player
		}
	}
	return nil
}
func FindWithDB_AccountId(id uint32) *TPlayer {
	if player, ok := g_account_cache[id]; ok {
		return player
	} else {
		player := new(TPlayer)
		if ok := player.LoadAllFromDB("accountid", id); ok {
			AddPlayerCache(player)
			return player
		}
	}
	return nil
}
func AddNewPlayer(accountId uint32, name string) *TPlayer {
	if player := NewPlayer(accountId, _GetNextPlayerID(), name); player != nil {
		AddPlayerCache(player)
		return player
	}
	return nil
}
func AddPlayerCache(player *TPlayer) {
	g_player_mutex.Lock()
	g_player_cache[player.Base.PlayerID] = player
	g_account_cache[player.Base.AccountID] = player
	g_player_mutex.Unlock()
}
func DelPlayerCache(playerId uint32) {
	if player, ok := g_player_cache[playerId]; ok {
		g_player_mutex.Lock()
		delete(g_player_cache, player.Base.PlayerID)
		delete(g_account_cache, player.Base.AccountID)
		g_player_mutex.Unlock()
	}
}
func _GetNextPlayerID() (ret uint32) {
	if g_auto_playerId == 0 {
		var lst []TBaseMoudle
		dbmgo.Find_Desc("Player", "_id", 1, &lst)
		if len(lst) > 0 {
			g_auto_playerId = lst[0].PlayerID + 1
		} else {
			g_auto_playerId = 10000
		}
	}
	g_player_mutex.Lock()
	ret = g_auto_playerId
	g_auto_playerId++
	g_player_mutex.Unlock()
	return
}
