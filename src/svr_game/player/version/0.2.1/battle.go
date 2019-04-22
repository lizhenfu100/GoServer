package player

const kDBBattle = "battle"

type TBattleModule struct {
	PlayerID uint32 `bson:"_id"`
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
}
