package player

import (
	"common"
	"dbmgo"
	"netConfig/meta"
	"svr_game/conf"
)

const KDBBattle = "battle"

type TBattleModule struct {
	PlayerID uint32 `bson:"_id"`
	Heros    map[uint8]THero
	Guns     map[uint16]TGun

	//TODO:zhoumf:玩家抽到的武器池子

	//小写私有数据，不入库
	owner *TPlayer //可能是nil
	aid   uint32
	name  string
	svrId int
}
type THero struct {
	ID    uint8
	Level uint8
	Exp   uint16
}
type TGun struct {
	ID    uint16
	Exp   uint16
	Level uint8
}

// ------------------------------------------------------------
// -- 框架接口
func (self *TBattleModule) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.Insert(KDBBattle, self)
	self._InitTempData(player)
}
func (self *TBattleModule) LoadFromDB(player *TPlayer) {
	if ok, _ := dbmgo.Find(KDBBattle, "_id", player.PlayerID, self); !ok {
		self.InitAndInsert(player)
	}
	self._InitTempData(player)
}
func (self *TBattleModule) WriteToDB() { dbmgo.UpdateId(KDBBattle, self.PlayerID, self) }
func (self *TBattleModule) OnLogin()   {}
func (self *TBattleModule) OnLogout()  {}
func (self *TBattleModule) _InitTempData(player *TPlayer) {
	if self.Heros == nil {
		self.Heros = make(map[uint8]THero)
	}
	self.owner = player
	self.aid = player.AccountID
	self.name = player.Name
	self.svrId = meta.G_Local.SvrID
}

// ------------------------------------------------------------
func (self *TBattleModule) BufToData(buf *common.NetPack) {
	self.aid = buf.ReadUInt32()
	self.name = buf.ReadString()
	self.svrId = buf.ReadInt()

	self._BufToData(buf)
}
func (self *TBattleModule) DataToBuf(buf *common.NetPack) {
	if self.owner != nil {
		buf.WriteUInt32(self.owner.AccountID)
		buf.WriteString(self.owner.Name)
	} else {
		buf.WriteUInt32(self.aid)
		buf.WriteString(self.name)
	}
	buf.WriteInt(self.svrId)

	self._DataToBuf(buf)
}
func (self *TBattleModule) _BufToData(buf *common.NetPack) {
	//英雄
	for cnt, i := buf.ReadUInt8(), uint8(0); i < cnt; i++ {
		id := buf.ReadUInt8()
		lv := buf.ReadUInt8()
		exp := buf.ReadUInt16()
		self.Heros[id] = THero{ID: id, Exp: exp, Level: lv}
	}
	//武器
	for cnt, i := buf.ReadUInt16(), uint16(0); i < cnt; i++ {
		id := buf.ReadUInt16()
		lv := buf.ReadUInt8()
		exp := buf.ReadUInt16()
		self.Guns[id] = TGun{ID: id, Exp: exp, Level: lv}
	}
}
func (self *TBattleModule) _DataToBuf(buf *common.NetPack) {
	//英雄数据
	buf.WriteUInt8(uint8(len(self.Heros)))
	for _, v := range self.Heros {
		buf.WriteUInt8(v.ID)
		buf.WriteUInt8(v.Level)
		buf.WriteUInt16(v.Exp)
	}
	//武器
	buf.WriteUInt16(uint16(len(self.Guns)))
	for _, v := range self.Guns {
		buf.WriteUInt16(v.ID)
		buf.WriteUInt8(v.Level)
		buf.WriteUInt16(v.Exp)
	}
}

// ------------------------------------------------------------
// --
func Rpc_game_write_db_battle_info(req, ack *common.NetPack, this *TPlayer) {
	this.battle._BufToData(req)
}
func Rpc_game_on_battle_end(req, ack *common.NetPack, this *TPlayer) {
	isWin := req.ReadBool()
	killCnt := req.ReadUInt8()   //击杀数
	assistCnt := req.ReadUInt8() //助攻数
	reviveCnt := req.ReadUInt8() //拉队友次数

	this.money.Add(KExp, 5)

	score := this.season.calcScore(isWin, killCnt, assistCnt, reviveCnt)
	this.season.AddScore(score)
}

// ------------------------------------------------------------
// -- 英雄
func (self *TBattleModule) AddHero(id uint8) bool {
	if _, ok := self.Heros[id]; ok {
		return false
	} else {
		self.Heros[id] = THero{ID: id}
		return true
	}
}
func (self *TBattleModule) AddHeroExp(id uint8, exp uint16) bool {
	kConf := conf.Const.Hero_LvUp
	if v, ok := self.Heros[id]; ok {
		if v.Level < uint8(len(kConf)) {
			if v.Exp += exp; v.Exp >= kConf[v.Level] {
				v.Exp -= kConf[v.Level]
				v.Level++
				self.Heros[id] = v
			}
			return true
		}
	}
	return false
}

// -- 武器
func (self *TBattleModule) AddGun(id uint16) bool {
	if _, ok := self.Guns[id]; ok {
		return false
	} else {
		self.Guns[id] = TGun{ID: id}
		return true
	}
}
func (self *TBattleModule) AddGunExp(id uint16, exp uint16) bool {
	kConf := conf.Const.Gun_LvUp
	if v, ok := self.Guns[id]; ok {
		if v.Level < uint8(len(kConf)) {
			if v.Exp += exp; v.Exp >= kConf[v.Level] {
				v.Exp -= kConf[v.Level]
				v.Level++
				self.Guns[id] = v
			}
			return true
		}
	}
	return false
}
