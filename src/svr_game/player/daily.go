package player

import (
	"common"
	"common/std"
	"dbmgo"
	"svr_game/conf"
	"time"
)

const kDBDaily = "daily"

type TDailyModule struct {
	PlayerID       uint32 `bson:"_id"`
	SignInYearWeek int    //签到，哪一周
	SignInMask     uint8  //周签到mask
}

// -------------------------------------
// -- 框架接口
func (self *TDailyModule) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.Insert(kDBDaily, self)
}
func (self *TDailyModule) LoadFromDB(player *TPlayer) {
	if ok, _ := dbmgo.Find(kDBDaily, "_id", player.PlayerID, self); !ok {
		self.InitAndInsert(player)
	}
}
func (self *TDailyModule) WriteToDB() { dbmgo.UpdateId(kDBDaily, self.PlayerID, self) }
func (self *TDailyModule) OnLogin() {
}
func (self *TDailyModule) OnLogout() {
}

// -------------------------------------
// -- API

// -------------------------------------
//! rpc
func Rpc_game_daily_sign_in(req, ack *common.NetPack, this *TPlayer) { //每日签到
	timenow := time.Now()
	wday := (timenow.Weekday() + 6) % 7 // weekday but Monday = 0.

	if int(wday) < len(conf.Const.DailySignInReward) && !std.GetBit8(this.daily.SignInMask, uint(wday)) {
		std.SetBit8(&this.daily.SignInMask, uint(wday), true)
		_, this.daily.SignInYearWeek = timenow.ISOWeek()

		reward := conf.Const.DailySignInReward[wday]
		length := len(reward)
		ack.WriteByte(byte(length))
		for i := 0; i < length; i++ {
			ack.WriteInt(reward[i].ID)
			ack.WriteInt(reward[i].Cnt)
		}
	} else {
		ack.WriteByte(0)
	}
}
func Rpc_game_look_over_daily_sign_in(req, ack *common.NetPack, this *TPlayer) {
	_, yearWeek := time.Now().ISOWeek()
	if this.daily.SignInYearWeek != yearWeek { //不是同一周了，清空签到记录
		this.daily.SignInMask = 0
	}
	rewards := conf.Const.DailySignInReward
	ack.WriteUInt8(this.daily.SignInMask)
	ack.WriteByte(byte(len(rewards)))
	for i := 0; i < len(rewards); i++ {
		reward := rewards[i]
		ack.WriteByte(byte(len(reward)))
		for i := 0; i < len(reward); i++ {
			ack.WriteInt(reward[i].ID)
			ack.WriteInt(reward[i].Cnt)
		}
	}
}
