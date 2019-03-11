package player

import (
	"common"
	"dbmgo"

	"svr_game/version/0.2.1"
)

const kDBBattle = "battle"

type TBattleModule struct {
	PlayerID uint32 `bson:"_id"`
	Diamond  uint32
	Exp      uint32
	Level    uint32
	Heros    []THeroInfo //英雄成长属性
}
type THeroInfo struct {
	HeroId uint8 //哪个英雄
	StarLv uint8 //升星
}

// ------------------------------------------------------------
func Upgrade(pid uint32) {
	ptr, newPtr := &TBattleModule{}, &player.TBattleModule{}
	if ok, _ := dbmgo.Find(kDBBattle, "_id", pid, ptr); ok {
		common.CopySameField(newPtr, ptr)
		dbmgo.UpdateIdSync(kDBBattle, pid, newPtr)
	}
}
