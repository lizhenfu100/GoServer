package player

import (
	"common"
	"conf"
	"dbmgo"
	"netConfig"
)

type TBattleModule struct {
	PlayerID uint32 `bson:"_id"`
	Diamond  uint32
	Exp      uint32
	Level    uint32
	Heros    []THeroInfo //英雄成长属性

	//小写私有数据，不入库
	owner *TPlayer
	aid   uint32
	name  string
	svrId int
}
type THeroInfo struct {
	HeroId    uint8 //哪个英雄
	StarLv    uint8 //升星
	IncHp     uint8 //增加的血量上限
	IncEnergy uint8 //增加的能量上限
	IncArmor  uint8 //增加的护甲上限
}

// ------------------------------------------------------------
// -- 框架接口
func (self *TBattleModule) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.InsertToDB("Battle", self)
	self._InitTempData(player)
}
func (self *TBattleModule) LoadFromDB(player *TPlayer) {
	if !dbmgo.Find("Battle", "_id", player.PlayerID, self) {
		self.InitAndInsert(player)
	}
	self._InitTempData(player)
}
func (self *TBattleModule) WriteToDB() { dbmgo.UpdateIdToDB("Battle", self.PlayerID, self) }
func (self *TBattleModule) OnLogin() {
}
func (self *TBattleModule) OnLogout() {
}
func (self *TBattleModule) _InitTempData(player *TPlayer) {
	self.owner = player
	self.aid = player.AccountID
	self.name = player.Name
	self.svrId = netConfig.G_Local_Meta.SvrID
}

// ------------------------------------------------------------
// -- API
func (self *TBattleModule) AddExp(exp int) {
	levelUpExp := uint32(0)
	if int(self.Level) < len(conf.SvrCsv.Exp_LvUp) {
		levelUpExp = conf.SvrCsv.Exp_LvUp[self.Level]
	} else {
		levelUpExp = conf.SvrCsv.Exp_Max
	}
	if self.Exp += uint32(exp); self.Exp >= levelUpExp {
		self.Exp -= levelUpExp
		self.Level++
	}
}

// ------------------------------------------------------------
// -- 打包玩家战斗数据
func (self *TBattleModule) BufToData(buf *common.NetPack) {
	self.aid = buf.ReadUInt32()
	self.name = buf.ReadString()
	self.svrId = buf.ReadInt()

	self._BufToDB(buf)
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

	self._DBToBuf(buf)
}
func (self *TBattleModule) _BufToDB(buf *common.NetPack) {
	self.Diamond = buf.ReadUInt32()
	self.Exp = buf.ReadUInt32()
	self.Level = buf.ReadUInt32()
	var v THeroInfo
	length := buf.ReadUInt8()
	self.Heros = self.Heros[:0]
	for i := uint8(0); i < length; i++ {
		v.HeroId = buf.ReadUInt8()
		v.StarLv = buf.ReadUInt8()
		v.IncHp = buf.ReadUInt8()
		v.IncEnergy = buf.ReadUInt8()
		v.IncArmor = buf.ReadUInt8()
		self.Heros = append(self.Heros, v)
	}
}
func (self *TBattleModule) _DBToBuf(buf *common.NetPack) {
	buf.WriteUInt32(self.Diamond)
	buf.WriteUInt32(self.Exp)
	buf.WriteUInt32(self.Level)
	buf.WriteUInt8(uint8(len(self.Heros)))
	for i := 0; i < len(self.Heros); i++ {
		ptr := &self.Heros[i]
		buf.WriteUInt8(ptr.HeroId)
		buf.WriteUInt8(ptr.StarLv)
		buf.WriteUInt8(ptr.IncHp)
		buf.WriteUInt8(ptr.IncEnergy)
		buf.WriteUInt8(ptr.IncArmor)
	}
}

// ------------------------------------------------------------
// --
func Rpc_game_write_db_battle_info(req, ack *common.NetPack, this *TPlayer) {
	this.Battle._BufToDB(req)
}
func Rpc_game_on_battle_end(req, ack *common.NetPack, this *TPlayer) {

}
