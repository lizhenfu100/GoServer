package player

import (
	"common"
	"conf"
	"dbmgo"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"nets/tcp"
	"sync"
	"time"
)

// ------------------------------------------------------------
// 封号，解封
func (self *TPlayer) Forbid() bool {
	if g_whitelist.Have(self.PlayerID) {
		return false
	} else {
		self.IsForbidden = true
		self.ForbidTime = time.Now().Unix()
		//TODO：强制踢下线
		return true
	}
}
func Rpc_game_permit_player(req, ack *common.NetPack, conn *tcp.TCPConn) {
	passwd := req.ReadString()
	pid := req.ReadUInt32()

	if passwd != conf.GM_Passwd {
		ack.WriteUInt16(err.Passwd_err)
	} else if p := GetPlayerBase(pid); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		p.IsForbidden = false
		dbmgo.UpdateId(kDBPlayer, p.PlayerID, bson.M{"$set": bson.M{
			"isforbidden": false}})
		ack.WriteUInt16(err.Success)
	}
	gamelog.Info("permit_player: %d", pid)
}

// ------------------------------------------------------------
// 白名单
var g_whitelist = &WhiteList{DBKey: "whitelist"}

type WhiteList struct {
	sync.Mutex `bson:"-"`
	DBKey      string `bson:"_id"`
	List       []uint32
}

func (self *WhiteList) Have(v uint32) bool {
	self.Lock()
	defer self.Unlock()
	for _, vv := range self.List {
		if vv == v {
			return true
		}
	}
	return false
}
func (self *WhiteList) Add(v uint32) {
	self.Lock()
	defer self.Unlock()
	for _, vv := range self.List {
		if vv == v {
			return
		}
	}
	self.List = append(self.List, v)
	self.UpdateDB()
}
func (self *WhiteList) Del(v uint32) {
	self.Lock()
	defer self.Unlock()
	for i, vv := range self.List {
		if vv == v {
			self.List = append(self.List[:i], self.List[i+1:]...)
			self.UpdateDB()
			return
		}
	}
}
func (self *WhiteList) UpdateDB() { dbmgo.UpdateId(dbmgo.KTableArgs, self.DBKey, self) }
func (self *WhiteList) InitDB() {
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", self.DBKey, self); !ok {
		dbmgo.Insert(dbmgo.KTableArgs, self)
	}
}

// ------------------------------------------------------------
func Rpc_game_whitelist_add(req, ack *common.NetPack, conn *tcp.TCPConn) {
	passwd := req.ReadString()
	pid := req.ReadUInt32()

	if passwd != conf.GM_Passwd {
		ack.WriteUInt16(err.Passwd_err)
	} else {
		g_whitelist.Add(pid)
		ack.WriteUInt16(err.Success)
	}
	gamelog.Info("whitelist_add: %d", pid)
}
func Rpc_game_whitelist_del(req, ack *common.NetPack, conn *tcp.TCPConn) {
	passwd := req.ReadString()
	pid := req.ReadUInt32()

	if passwd != conf.GM_Passwd {
		ack.WriteUInt16(err.Passwd_err)
	} else {
		g_whitelist.Del(pid)
		ack.WriteUInt16(err.Success)
	}
	gamelog.Info("whitelist_del: %d", pid)
}
