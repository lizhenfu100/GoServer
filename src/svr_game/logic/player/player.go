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
	"common/service"
	"dbmgo"
	"gamelog"
	"sync/atomic"
	"time"
)

var (
	G_ServiceMgr service.ServiceMgr
)

const (
	Idle_Max_Second     = 60
	ReLogin_Wait_Second = time.Second * 120

	//须与ServiceMgr初始化顺序一致
	Service_Write_DB  = 0
	Service_Check_AFK = 1
)

func init() {
	G_ServiceMgr = service.ServiceMgr{[]service.IService{
		service.NewServicePatch(_Service_Write_DB, 15*60*1000),
		service.NewServiceVec(_Service_Check_AFK, 1000),
	}}
}

type iMoudle interface {
	InitAndInsert(*TPlayer)
	LoadFromDB(*TPlayer)
	WriteToDB()
	OnLogin()
	OnLogout()
}
type TPlayerBase struct {
	PlayerID   uint32 `bson:"_id"`
	AccountID  uint32
	Name       string
	LoginTime  int64
	LogoutTime int64
}
type TPlayer struct {
	//temp data
	moudles   []iMoudle
	askchan   chan func(*TPlayer)
	_isOnlnie int32
	_idleSec  uint32
	pTeam     *TeamData
	//db data
	TPlayerBase
	Mail   TMailMoudle
	Friend TFriendMoudle
	Chat   TChatMoudle
	Battle TBattleMoudle
	Save   TSaveClient
}

func _NewPlayer() *TPlayer {
	player := new(TPlayer)
	//! regist
	player.moudles = []iMoudle{
		&player.Mail,
		&player.Friend,
		&player.Chat,
		&player.Battle,
		&player.Save,
	}
	player.askchan = make(chan func(*TPlayer), 128)
	return player
}
func _NewPlayerInDB(accountId uint32, id uint32, name string) *TPlayer {
	player := _NewPlayer()
	// if dbmgo.Find("Player", "name", name, player) { //禁止重名
	// 	return nil
	// }
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
	player := _NewPlayer()
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
func (self *TPlayer) Login() {
	atomic.StoreInt32(&self._isOnlnie, 1)
	self.LoginTime = time.Now().Unix()
	atomic.SwapUint32(&self._idleSec, 0)
	for _, v := range self.moudles {
		v.OnLogin()
	}

	G_ServiceMgr.Register(Service_Write_DB, self)
	G_ServiceMgr.Register(Service_Check_AFK, self)
}
func (self *TPlayer) Logout() {
	atomic.StoreInt32(&self._isOnlnie, 0)
	self.LogoutTime = time.Now().Unix()
	for _, v := range self.moudles {
		v.OnLogout()
	}
	self.ExitTeam()

	G_ServiceMgr.UnRegister(Service_Write_DB, self)
	G_ServiceMgr.UnRegister(Service_Check_AFK, self)

	//Notice: AfterFunc是在另一线程执行，所以传入函数必须线程安全
	time.AfterFunc(ReLogin_Wait_Second, func() {
		// 延时删除，提升重连效率
		if !self.IsOnline() && FindPlayerInCache(self.PlayerID) != nil {
			gamelog.Debug("Pid(%d) Delete", self.PlayerID)
			go self.WriteAllToDB()
			DelPlayerCache(self)
		}
	})
}
func (self *TPlayer) IsOnline() bool { return atomic.LoadInt32(&self._isOnlnie) > 0 }

// -------------------------------------
// service
func _Service_Write_DB(ptr interface{}) {
	if player, ok := ptr.(*TPlayer); ok {
		player.WriteAllToDB()
	}
}
func _Service_Check_AFK(ptr interface{}) {
	if player, ok := ptr.(*TPlayer); ok && player.IsOnline() {
		atomic.AddUint32(&player._idleSec, 1)
		if atomic.LoadUint32(&player._idleSec) > Idle_Max_Second {
			gamelog.Debug("Pid(%d) AFK", player.PlayerID)
			player.Logout()
		}
	}
}

// ------------------------------------------------------------
//! for other player write my data
func AsyncNotifyPlayer(pid uint32, handler func(*TPlayer)) {
	if player := FindPlayerInCache(pid); player != nil {
		player.AsyncNotify(handler)
	}
}
func (self *TPlayer) AsyncNotify(handler func(*TPlayer)) {
	if self.IsOnline() {
		select {
		case self.askchan <- handler:
		default:
			gamelog.Warn("Player askChan is full !!!")
			return
		}
	} else { //TODO:zhoumf: 如何安全方便的修改离线玩家数据

		//准备将离线的操作转给mainloop，这样所有离线玩家就都在一个chan里处理了
		//要是中途玩家上线，mainloop的chan里还有他的操作没处理完怎么整！？囧~
		//mainloop设计成map<pid, chan>，玩家上线时，检测自己的chan有效否，等它处理完？

		//gen_server
		//将某个独立模块的所有操作扔进gen_server，外界只读(有滞后性)
		//会加大代码量，每个操作都得转一次到chan
		//【Notice】可能gen_server里还有修改操作，且玩家已下线，会重新读到内存，此时修改完毕后须及时入库

		//设计统一的接口，编辑离线数据，也很麻烦呐
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

// ------------------------------------------------------------
//! 访问玩家部分数据，包括离线的
func GetPlayerBaseData(pid uint32) *TPlayerBase {
	if player := FindPlayerInCache(pid); player != nil {
		return &player.TPlayerBase
	} else {
		ptr := new(TPlayerBase)
		if dbmgo.Find("Player", "_id", pid, ptr) {
			return ptr
		}
		return nil
	}
}
