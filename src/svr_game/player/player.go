/***********************************************************************
* @ 玩家数据
* @ brief
	1、数据散列模块化，按业务区分成块，各自独立处理，见iModule
	2、可调用DB【同步读单个模块】，编辑后再【异步写】

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
	"netConfig/meta"
	"svr_game/version"
	"sync/atomic"
	"tcp"
	"time"
)

const (
	Idle_Max_Minute   = 3 //须客户端心跳包
	ReLogin_Wait_Time = time.Minute * 5
	kDBPlayer         = "Player"
)

type TPlayerBase struct {
	PlayerID   uint32 `bson:"_id"`
	AccountID  uint32 //用于网络通信：一个账号下可能有多个角色，但仅可能一个在线
	LoginTime  int64
	LogoutTime int64
	Name       string
	Head       string
	Version    string //用于数据升级；客户端连接与自己版本匹配的节点
}
type iModule interface {
	InitAndInsert(*TPlayer)
	LoadFromDB(*TPlayer)
	WriteToDB()
	OnLogin()
	OnLogout()
}
type TPlayer struct {
	_isOnlnie int32
	_idleMin  uint32 //每次收到消息时归零
	conn      *tcp.TCPConn
	//askchan   chan func(*TPlayer)

	/* --- db data --- */
	TPlayerBase
	modules []iModule
	mail    TMailModule
	friend  TFriendModule
	team    Team
	battle  TBattleModule
	season  TSeasonModule
}

func _NewPlayer() *TPlayer {
	self := new(TPlayer)
	self.init()
	return self
}
func (self *TPlayer) init() {
	//self.askchan = make(chan func(*TPlayer), 128)
	self.modules = []iModule{ //regist
		&self.mail,
		&self.friend,
		&self.team,
		&self.battle,
		&self.season,
	}
}

// -------------------------------------
// modules
func NewPlayerInDB(accountId uint32, name string) *TPlayer {
	player := _NewPlayer()
	// if dbmgo.Find("Player", "name", name, player) { //禁止重名
	// 	return nil
	// }
	player.Name = name
	player.AccountID = accountId
	player.PlayerID = accountId //一个账户下仅一个角色的游戏，可令pid=aid
	//player.PlayerID = dbmgo.GetNextIncId("PlayerId")
	player.Version = meta.G_Local.Version

	if dbmgo.InsertSync(kDBPlayer, &player.TPlayerBase) {
		for _, v := range player.modules {
			v.InitAndInsert(player)
		}
		AddCache(player)
		return player
	}
	return nil
}
func LoadPlayerFromDB(key string, val uint32) *TPlayer {
	player := _NewPlayer()
	if ok, _ := dbmgo.Find(kDBPlayer, key, val, &player.TPlayerBase); ok {
		if player.Version != meta.G_Local.Version {
			version.Upgrade(player.PlayerID, player.Version, meta.G_Local.Version)
		}
		for _, v := range player.modules {
			v.LoadFromDB(player)
		}
		AddCache(player)
		return player
	}
	return nil
}
func (self *TPlayer) WriteAllToDB() {
	dbmgo.UpdateId(kDBPlayer, self.PlayerID, &self.TPlayerBase)
	for _, v := range self.modules {
		v.WriteToDB()
	}
}
func (self *TPlayer) Login(conn *tcp.TCPConn) {
	if atomic.SwapInt32(&self._isOnlnie, 1) == 0 {
		atomic.AddInt32(&g_online_cnt, 1)
	}
	atomic.StoreUint32(&self._idleMin, 0)
	atomic.StoreInt64(&self.LoginTime, time.Now().Unix())
	self.conn = conn
	if conn != nil && conn.UserPtr == nil { //链接可能是gateway节点
		conn.UserPtr = self
	}
	for _, v := range self.modules {
		v.OnLogin()
	}

	G_ServiceMgr.Register(Service_Write_DB, self)
	G_ServiceMgr.Register(Service_Check_AFK, self)
}
func (self *TPlayer) Logout() {
	if atomic.SwapInt32(&self._isOnlnie, 0) > 0 {
		atomic.AddInt32(&g_online_cnt, -1)
	}
	atomic.StoreInt64(&self.LogoutTime, time.Now().Unix())
	for _, v := range self.modules {
		v.OnLogout()
	}
	self.WriteAllToDB()

	G_ServiceMgr.UnRegister(Service_Write_DB, self)
	G_ServiceMgr.UnRegister(Service_Check_AFK, self)

	//Notice: AfterFunc是在另一线程执行，所以传入函数必须线程安全
	time.AfterFunc(ReLogin_Wait_Time, func() {
		// 延时删除，提升重连效率
		if !self.IsOnline() && FindAccountId(self.AccountID) != nil {
			gamelog.Debug("Aid(%d) Delete", self.AccountID)
			self.WriteAllToDB()
			DelCache(self)
		}
	})
}
func (self *TPlayer) IsOnline() bool { return atomic.LoadInt32(&self._isOnlnie) > 0 }

// -------------------------------------
// service
const ( //须与ServiceMgr初始化顺序一致
	Service_Write_DB  = 0
	Service_Check_AFK = 1
)

var G_ServiceMgr service.ServiceMgr

func init() {
	G_ServiceMgr = service.ServiceMgr{
		Chan: make(chan service.Obj, 2048),
		List: []service.IService{
			service.NewServicePatch(_Service_Write_DB, 15*60*1000),
			service.NewServiceVec(_Service_Check_AFK, 60*1000),
		},
	}
}
func _Service_Write_DB(ptr interface{}) {
	if player, ok := ptr.(*TPlayer); ok {
		player.WriteAllToDB()
	}
}
func _Service_Check_AFK(ptr interface{}) {
	if player, ok := ptr.(*TPlayer); ok && player.IsOnline() {
		if atomic.AddUint32(&player._idleMin, 1) > Idle_Max_Minute {
			gamelog.Debug("Aid(%d) AFK", player.AccountID)
			player.Logout()
		}
	}
}
