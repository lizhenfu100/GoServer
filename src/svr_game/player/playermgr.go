package player

import (
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"svr_game/logic/season"
	"sync"
	"time"
)

var (
	G_players    sync.Map //<pid, *TPlayer>
	g_accounts   sync.Map //<aid, *TPlayer>
	g_online_cnt int32
)

func InitDB() {
	InitSvrMailDB()
	season.InitDB()
	g_whitelist.InitDB()

	var list1 []TPlayerBase //只载入近期登录过的
	dbmgo.FindAll(kDBPlayer, bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - 24*3600}}, &list1)
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
}

//! 若多线程架构，玩家内存，只能他自己直接修改，别人须转给他后间接改(异步)
func FindPid(pid uint32) *TPlayer {
	if v, ok := G_players.Load(pid); ok {
		return v.(*TPlayer)
	}
	return nil
}
func FindAccountId(aid uint32) *TPlayer {
	if v, ok := g_accounts.Load(aid); ok {
		return v.(*TPlayer)
	}
	return nil
}
func FindWithDB(pid uint32) *TPlayer {
	if player := FindPid(pid); player != nil {
		return player
	} else {
		return LoadPlayerFromDB("_id", pid)
	}
}

// -------------------------------------
//! 辅助函数
func AddCache(player *TPlayer) {
	G_players.Store(player.PlayerID, player)
	g_accounts.Store(player.AccountID, player)
}
func DelCache(player *TPlayer) {
	G_players.Delete(player.PlayerID)
	g_accounts.Delete(player.AccountID)
}

// ------------------------------------------------------------
//! 访问玩家部分数据，包括离线的
func GetPlayerBase(pid uint32) *TPlayerBase {
	if player := FindPid(pid); player != nil {
		return &player.TPlayerBase
	} else {
		ptr := new(TPlayerBase)
		if ok, _ := dbmgo.Find(kDBPlayer, "_id", pid, ptr); ok {
			return ptr
		}
		return nil
	}
}
