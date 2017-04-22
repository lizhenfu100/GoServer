package player

import (
	"sync"
)

var (
	g_player_mutex sync.Mutex
	g_player_cache map[uint32]*TPlayer
)

func FindPlayerInCache(id uint32) *TPlayer { return g_player_cache[id] }
func FindPlayerWithDB(id uint32) *TPlayer {
	if player, ok := g_player_cache[id]; ok {
		return player
	} else {
		player := new(TPlayer)
		if ok := player.LoadAllFromDB(id); ok {
			g_player_mutex.Lock()
			g_player_cache[id] = player
			g_player_mutex.Unlock()
			return player
		}
	}
	return nil
}
func AddNewPlayer(accountId uint32, id uint32, name string) *TPlayer {
	if _, ok := g_player_cache[id]; ok == false {
		if player := NewPlayer(accountId, id, name); player != nil {
			g_player_mutex.Lock()
			g_player_cache[id] = player
			g_player_mutex.Unlock()
			return player
		}
	}
	return nil
}
func DeletePlayerCache(id uint32) {
	g_player_mutex.Lock()
	delete(g_player_cache, id)
	g_player_mutex.Unlock()
}
