package player

import (
	"sync"
)

var (
	g_player_mutex sync.Mutex
	g_player_lst   map[uint32]*TPlayer
)

func GetPlayerOnline(id uint32) *TPlayer { return g_player_lst[id] }
func FindPlayerWithDB(id uint32) *TPlayer {
	if player, ok := g_player_lst[id]; ok {
		return player
	} else {
		player := new(TPlayer)
		if ok := player.LoadAllFromDB(id); ok {
			g_player_mutex.Lock()
			g_player_lst[id] = player
			g_player_mutex.Unlock()
			return player
		}
	}
	return nil
}
func AddNewPlayer(accountId uint32, id uint32, name string) *TPlayer {
	if _, ok := g_player_lst[id]; ok == false {
		if player := NewPlayer(accountId, id, name); player != nil {
			g_player_mutex.Lock()
			g_player_lst[id] = player
			g_player_mutex.Unlock()
			return player
		}
	}
	return nil
}
func DeletePlayer(id uint32) {
	g_player_mutex.Lock()
	delete(g_player_lst, id)
	g_player_mutex.Unlock()
}
