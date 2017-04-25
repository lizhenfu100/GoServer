/***********************************************************************
* @ 玩家数据
* @ brief
	1、数据散列模块化，按业务区分成块，各自独立处理，如：TBaseMoudle、TMailMoudle
	2、可调用DB【同步读单个模块】，编辑后再【异步写】
	3、本文件的数据库操作接口，都是【同步的】

* @ 访问离线玩家
	1、用什么取什么，读出一块数据编辑完后写回，尽量少载入整个玩家结构体
	2、设想把TPlayer里的数据块部分全定义为指针，各模块分别做个缓存表(online list、offline list)
	3、但觉得有些设计冗余，缓存这种事情，应该交给DBCache系统做，业务层不该负责这事儿

* @ 自动写数据库
	1、借助ServicePatch，十五分钟全写一遍在线玩家，重要数据才手动异步写dbmgo.InsertToDB
	2、关服，须先踢所有玩家下线，触发Logou流程写库，再才能关闭进程

* @ author zhoumf
* @ date 2017-4-22
***********************************************************************/
package player

import (
	"common"
	"dbmgo"
	"sync"

	"svr_game/logic/mail"
)

var (
	G_auto_write_db = common.NewServicePatch(_WritePlayerToDB, 15*60*1000)
)

type PlayerMoudle interface {
	InitAndInsert(id uint32)
	LoadFromDB(id uint32)
	WriteToDB()
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
		player.InitAndInsertDB()
		return player
	}
	return nil
}
func (self *TPlayer) InitAndInsertDB() {
	for _, v := range self.moudles {
		v.InitAndInsert(self.Base.PlayerID)
	}
}
func (self *TPlayer) LoadAllFromDB(key string, val uint32) bool {
	if ok := dbmgo.Find("Player", key, val, &self.Base); ok {
		for _, v := range self.moudles {
			v.LoadFromDB(self.Base.PlayerID)
		}
		return true
	}
	return false
}
func (self *TPlayer) WriteAllToDB() {
	dbmgo.UpdateSync("Player", self.Base.PlayerID, &self.Base)
	for _, v := range self.moudles {
		v.WriteToDB()
	}
}
func (self *TPlayer) OnLogin() {
	for _, v := range self.moudles {
		v.OnLogin()
	}
	G_auto_write_db.Register(self)
}
func (self *TPlayer) OnLogout() {
	for _, v := range self.moudles {
		v.OnLogout()
	}
	G_auto_write_db.UnRegister(self)
}
func _WritePlayerToDB(ptr interface{}) {
	if player, ok := ptr.(TPlayer); ok {
		player.WriteAllToDB()
	}
}
