/***********************************************************************
* @ 玩家数据
* @ brief
	1、数据散列模块化，按业务区分成块，各自独立处理，如：TMailMoudle
	2、可调用DB【同步读单个模块】，编辑后再【异步写】
	3、本文件的数据库操作接口，都是【同步的】

* @ 访问离线玩家
	1、用什么取什么，读出一块数据编辑完后写回，尽量少载入整个玩家结构体
	2、设想把TPlayer里的数据块部分全定义为指针，各模块分别做个缓存表(online list、offline list)
	3、但觉得有些设计冗余，缓存这种事情，应该交给DBCache系统做，业务层不该负责这事儿

* @ 自动写数据库
	1、借助ServicePatch，十五分钟全写一遍在线玩家，重要数据才手动异步写dbmgo.InsertToDB
	2、关服，须先踢所有玩家下线，触发Logou流程写库，再才能关闭进程

* @ 玩家之间互改数据【多线程架构】
	1、禁止直接操作对方内存

	2、异步间接改别人的数据
			*、提供统一接口，将写操作发送到目标所在线程，让目标自己改写
			*、因为读别人数据是直接拿内存，此方式可能带来时序Bug【异步写在读之前，但读到旧数据】
			*、比如：异步扣别人100块，又立即读，可能他还是没钱

	3、分别加读写锁【多读少写用RWMutex，写也多的用Mutex】
			*、会被其他人改的数据块，性质上同全局数据类似，多读少写的
			*、读写锁封装接口，谁都不允许直接访问
			*、比异步方式(可能读到旧值)安全，但要写好锁代码【屏蔽所有竞态条件、无死锁】可不是件容易事~_~

* @ author zhoumf
* @ date 2017-4-22
***********************************************************************/
package player

import (
	"common"
	"dbmgo"
	"gamelog"
	"time"
)

var (
	G_Auto_Write_DB = common.NewServicePatch(_WritePlayerToDB, 15*60*1000)
)

type TPlayer struct {
	//db data
	TPlayerBase
	Mail   TMailMoudle
	Friend TFriendMoudle
	Chat   TChatMoudle
	//temp data
	moudles []PlayerMoudle
	askchan chan func(*TPlayer)
	isOnlie bool
}
type TPlayerBase struct {
	PlayerID   uint32 `bson:"_id"`
	AccountID  uint32
	Name       string
	LoginTime  int64
	LogoutTime int64
}
type PlayerMoudle interface {
	InitAndInsert(*TPlayer)
	LoadFromDB(*TPlayer)
	WriteToDB()
	OnLogin()
	OnLogout()
}

func NewPlayer() *TPlayer {
	player := new(TPlayer)
	//! regist
	player.moudles = []PlayerMoudle{
		&player.Mail,
		&player.Friend,
		&player.Chat,
	}
	player.askchan = make(chan func(*TPlayer), 128)
	return player
}
func NewPlayerInDB(accountId uint32, id uint32, name string) *TPlayer {
	player := NewPlayer()
	player.AccountID = accountId
	player.PlayerID = id
	player.Name = name
	if dbmgo.InsertSync("Player", &player.TPlayerBase) {
		for _, v := range player.moudles {
			v.InitAndInsert(player)
		}
		return player
	}
	return nil
}
func LoadPlayerFromDB(key string, val uint32) *TPlayer {
	player := NewPlayer()
	if dbmgo.Find("Player", key, val, &player.TPlayerBase) {
		for _, v := range player.moudles {
			v.LoadFromDB(player)
		}
		return player
	}
	return nil
}
func (self *TPlayer) WriteAllToDB() {
	if dbmgo.UpdateSync("Player", self.PlayerID, &self.TPlayerBase) {
		for _, v := range self.moudles {
			v.WriteToDB()
		}
	}
}
func (self *TPlayer) OnLogin() {
	self.isOnlie = true
	for _, v := range self.moudles {
		v.OnLogin()
	}
	G_Auto_Write_DB.Register(self)
}
func (self *TPlayer) OnLogout() {
	self.isOnlie = false
	for _, v := range self.moudles {
		v.OnLogout()
	}
	G_Auto_Write_DB.UnRegister(self)

	// 延时30s后再删，提升重连效率
	time.AfterFunc(30*time.Second, func() {
		if !self.isOnlie {
			go self.WriteAllToDB()
			DelPlayerCache(self.PlayerID)
		}
	})
}
func _WritePlayerToDB(ptr interface{}) {
	if player, ok := ptr.(TPlayer); ok {
		player.WriteAllToDB()
	}
}

//! for other player write my data
func AsyncNotifyPlayer(id uint32, handler func(*TPlayer)) {
	if player := _FindPlayerInCache(id); player != nil {
		select {
		case player.askchan <- handler:
		default:
			gamelog.Warn("Player askChan is full !!!")
			return
		}
	}
}
func (self *TPlayer) _HandleAsyncNotify() {
	for {
		select {
		case handler := <-self.askchan:
			handler(self)
		default:
			return
		}
	}
}
