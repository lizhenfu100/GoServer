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
	Heros    []THero
	Guns     []TGun

	//TODO:zhoumf:玩家抽到的武器池子

	//小写私有数据，不入库
	owner *TPlayer //可能是nil
	show  TShowInfo
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
	self.owner = player
	self.show = *player.GetShowInfo()
	self.svrId = meta.G_Local.SvrID
}

// ------------------------------------------------------------
// 客户端：ShowInfo.cs  战斗服：CrossAgent.56
func (self *TBattleModule) BufToData(buf *common.NetPack) {
	self.show.BufToData(buf)
	self.svrId = buf.ReadInt()
	self._BufToData(buf)
}
func (self *TBattleModule) DataToBuf(buf *common.NetPack) {
	//1. show info
	if self.owner != nil {
		self.owner.GetShowInfo().DataToBuf(buf)
	} else {
		self.show.DataToBuf(buf)
	}
	//2. svr_game id
	buf.WriteInt(self.svrId)
	//3. battle data
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
func (self *TBattleModule) GetHero(id uint8) *THero {
	for i := 0; i < len(self.Heros); i++ {
		if self.Heros[i].ID == id {
			return &self.Heros[i]
		}
	}
	return nil
}
func (self *TBattleModule) AddHero(id uint8) bool {
	if self.GetHero(id) == nil {
		self.Heros = append(self.Heros, THero{ID: id})
		return true
	}
	return false
}
func (self *TBattleModule) AddHeroExp(id uint8, exp uint16) bool {
	kConf := conf.Const.Hero_LvUp
	if p := self.GetHero(id); p != nil {
		if p.Level < uint8(len(kConf)) {
			if p.Exp += exp; p.Exp >= kConf[p.Level] {
				p.Exp -= kConf[p.Level]
				p.Level++
			}
			return true
		}
	}
	return false
}

// -- 武器
func (self *TBattleModule) GetGun(id uint16) *TGun {
	for i := 0; i < len(self.Guns); i++ {
		if self.Guns[i].ID == id {
			return &self.Guns[i]
		}
	}
	return nil
}
func (self *TBattleModule) AddGun(id uint16) bool {
	if self.GetGun(id) == nil {
		self.Guns = append(self.Guns, TGun{ID: id})
		return true
	}
	return false
}
func (self *TBattleModule) AddGunExp(id uint16, exp uint16) bool {
	kConf := conf.Const.Gun_LvUp
	if p := self.GetGun(id); p != nil {
		if p.Level < uint8(len(kConf)) {
			if p.Exp += exp; p.Exp >= kConf[p.Level] {
				p.Exp -= kConf[p.Level]
				p.Level++
			}
			return true
		}
	}
	return false
}
