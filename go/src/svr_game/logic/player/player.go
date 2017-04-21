package player

import (
	"dbmgo"
	"sync"

	"svr_game/logic/mail"
)

type PlayerMoudle interface {
	InitWriteDB(id uint32)
	LoadFromDB(id uint32)
	OnLogin()
	OnLogout()
}
type TBaseMoudle struct {
	PlayerID   uint32 `bson:"_id"`
	AccountID  uint32
	Name       string
	LoginTime  int64
	LogoutTime int64
}
type TPlayer struct {
	//db data
	Base TBaseMoudle
	Mail mail.TMailMoudle
	//temp data
	mutex   sync.Mutex
	moudles []PlayerMoudle
}

func NewPlayer(accountId uint32, id uint32, name string) *TPlayer {
	player := new(TPlayer)
	//! regist
	player.moudles = []PlayerMoudle{
		&player.Mail,
	}
	player.Base.AccountID = accountId
	player.Base.PlayerID = id
	player.Base.Name = name
	if err := dbmgo.InsertSync("Player", &player.Base); err != nil {
		player.InitWriteDB()
		return player
	}
	return nil
}
func (self *TPlayer) InitWriteDB() {
	for _, v := range self.moudles {
		v.InitWriteDB(self.Base.PlayerID)
	}
}
func (self *TPlayer) LoadAllFromDB(id uint32) bool {
	if ok := dbmgo.Find("Player", "_id", id, &self.Base); ok {
		for _, v := range self.moudles {
			v.LoadFromDB(id)
		}
		return true
	}
	return false
}
func (self *TPlayer) OnLogin() {
	for _, v := range self.moudles {
		v.OnLogin()
	}
}
func (self *TPlayer) OnLogout() {
	for _, v := range self.moudles {
		v.OnLogout()
	}
}
