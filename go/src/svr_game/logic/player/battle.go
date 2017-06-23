/***********************************************************************
* @ 战斗匹配
* @ brief
	*、玩法类型
		1、球球类型的，允许中途加入房间，匹配得送到战斗服做
		2、王者这样的，匹配直接在GameSvr做，组好一场人，全送进某个战斗服即可

* @ author zhoumf
* @ date 2017-6-5
***********************************************************************/
package player

import (
	"common"
	"dbmgo"
	"svr_game/api"
	"tcp"
	// "gopkg.in/mgo.v2/bson"
)

type TBattleMoudle struct {
	PlayerID uint32 `bson:"_id"`

	owner          *TPlayer
	loginBattleMsg *common.ByteBuffer
	isShowWaitUI   bool
}

// -------------------------------------
// -- 框架接口
func (self *TBattleMoudle) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	self.owner = player
	dbmgo.InsertSync("Battle", self)

	self._InitTempData()
}
func (self *TBattleMoudle) WriteToDB() { dbmgo.UpdateSync("Battle", self.PlayerID, self) }
func (self *TBattleMoudle) LoadFromDB(player *TPlayer) {
	dbmgo.Find("Battle", "_id", player.PlayerID, self)
	self.owner = player

	self._InitTempData()
}
func (self *TBattleMoudle) OnLogin() {
}
func (self *TBattleMoudle) OnLogout() {
}
func (self *TBattleMoudle) _InitTempData() {
	self.loginBattleMsg = common.NewByteBufferCap(32)
}

// -------------------------------------
// -- API
func Rpc_Battle_Begin(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	if self.pTeam == nil || self.pTeam.lst[0] != self {
		return
	}
	battleMsg := common.NewNetPackCap(32)
	battleMsg.WriteByte(byte(len(self.pTeam.lst)))
	for _, ptr := range self.pTeam.lst {
		battleMsg.WriteUInt32(ptr.PlayerID)
		// pack player battle data
		battleMsg.WriteString(ptr.Name)

		//重新匹配战斗服
		ptr.Battle.loginBattleMsg.Clear()

		// 通知队员，开等待界面
		if ptr != self {
			ptr.AsyncNotify(func(p *TPlayer) {
				p.Battle.isShowWaitUI = true
			})
		}
	}
	battleMsg.SetRpc("rpc_cross_relay_battle_data")
	api.SendToCross(battleMsg)
}
func Rpc_Battle_Ack(req, ack *common.NetPack, conn *tcp.TCPConn) {
	battleSvrIP := req.ReadString()
	battleSvrPort := req.ReadUInt16()
	cnt := req.ReadByte()
	for i := byte(0); i < cnt; i++ {
		pid := req.ReadUInt32()
		//通知client登录战斗服
		AsyncNotifyPlayer(pid, func(player *TPlayer) {
			buf := player.Battle.loginBattleMsg
			buf.WriteString(battleSvrIP)
			buf.WriteUInt16(battleSvrPort)
		})
	}
}
func Rpc_Probe_Login_Battle(req, ack *common.NetPack, ptr interface{}) {
	player := ptr.(*TPlayer)

	msg := player.Battle.loginBattleMsg
	if msg.Size() > 0 {
		ack.WriteInt8(1)
		ack.WriteBuf(msg.DataPtr)
		player.Battle.loginBattleMsg.Clear()
	} else {
		ack.WriteInt8(-1)
	}
}
