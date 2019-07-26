package player

import "dbmgo"

const (
	KDBMoney = "money"
	//货币类型
	KDiamond = "diamond"
	KExp     = "exp"
)

type TMoneyModule struct {
	PlayerID uint32 `bson:"_id"`
	Token    map[string]int

	//小写私有数据，不入库
	owner *TPlayer
}

// ------------------------------------------------------------
// -- 框架接口
func (self *TMoneyModule) InitAndInsert(p *TPlayer) {
	self.PlayerID = p.PlayerID
	dbmgo.Insert(KDBMoney, self)
	self._InitTempData(p)
}
func (self *TMoneyModule) LoadFromDB(p *TPlayer) {
	if ok, _ := dbmgo.Find(KDBMoney, "_id", p.PlayerID, self); !ok {
		self.InitAndInsert(p)
	}
	self._InitTempData(p)
}
func (self *TMoneyModule) WriteToDB() { dbmgo.UpdateId(KDBMoney, self.PlayerID, self) }
func (self *TMoneyModule) OnLogin() {
}
func (self *TMoneyModule) OnLogout() {
}
func (self *TMoneyModule) _InitTempData(p *TPlayer) {
	self.owner = p
}

// ------------------------------------------------------------
// -- API
func (self *TMoneyModule) Add(typ string, n int) {
	if n > 0 {
		v := self.Token[typ] + n
		self.Token[typ] = v
	}
}
func (self *TMoneyModule) Del(typ string, n int) bool {
	if n <= 0 {
		return false
	}
	if v := self.Token[typ] - n; v >= 0 {
		self.Token[typ] = v
		return true
	} else {
		return false
	}
}
