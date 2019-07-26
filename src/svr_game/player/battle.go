package player

import (
	"common"
	"dbmgo"
	"netConfig/meta"
	"svr_game/conf"
)

const kDBBattle = "battle"

type TBattleModule struct {
	PlayerID uint32 `bson:"_id"`
	Heros    map[uint8]THero

	//小写私有数据，不入库
	owner *TPlayer //可能是nil
	aid   uint32
	name  string
	svrId int
}
type THero struct {
	ID    uint8
	Level uint8 //1起始
	Exp   uint16
}

// ------------------------------------------------------------
// -- 框架接口
func (self *TBattleModule) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.Insert(kDBBattle, self)
	self._InitTempData(player)
}
func (self *TBattleModule) LoadFromDB(player *TPlayer) {
	if ok, _ := dbmgo.Find(kDBBattle, "_id", player.PlayerID, self); !ok {
		self.InitAndInsert(player)
	}
	self._InitTempData(player)
}
func (self *TBattleModule) WriteToDB() { dbmgo.UpdateId(kDBBattle, self.PlayerID, self) }
func (self *TBattleModule) OnLogin() {
}
func (self *TBattleModule) OnLogout() {
}
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
// -- API
func (self *TBattleModule) AddHero(id uint8) *THero {
	if _, ok := self.Heros[id]; ok {
		return nil
	} else {
		v := THero{id, 1, 0}
		self.Heros[id] = v
		return &v
	}
}
func (self *TBattleModule) GetHero(id uint8) *THero {
	if v, ok := self.Heros[id]; ok {
		return &v
	}
	return nil
}
func (self *TBattleModule) AddHeroExp(id uint8, exp uint16) bool {
	if p := self.GetHero(id); p != nil {
		if kConf := conf.Const.Hero_LvUp; p.Level < uint8(len(kConf)) {
			if p.Exp += exp; p.Exp >= kConf[p.Level] {
				p.Exp -= kConf[p.Level]
				p.Level++
			}
			return true
		}
	}
	return false
}

// ------------------------------------------------------------
// -- 打包玩家战斗数据
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
	for cnt, i := buf.ReadUInt8(), uint8(0); i < cnt; i++ {
		id := buf.ReadUInt8()
		lv := buf.ReadUInt8()
		exp := buf.ReadUInt16()
		self.Heros[id] = THero{id, lv, exp}
	}
}
func (self *TBattleModule) _DataToBuf(buf *common.NetPack) {
	buf.WriteUInt8(uint8(len(self.Heros)))
	for _, v := range self.Heros {
		buf.WriteUInt8(v.ID)
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
