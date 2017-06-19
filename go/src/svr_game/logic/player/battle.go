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
}

// -------------------------------------
// -- 框架接口
func (self *TBattleMoudle) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	self.owner = player
	dbmgo.InsertSync("Battle", self)

	self.loginBattleMsg = common.NewByteBufferCap(32)
}
func (self *TBattleMoudle) WriteToDB() { dbmgo.UpdateSync("Battle", self.PlayerID, self) }
func (self *TBattleMoudle) LoadFromDB(player *TPlayer) {
	dbmgo.Find("Battle", "_id", player.PlayerID, self)
	self.owner = player

	self.loginBattleMsg = common.NewByteBufferCap(32)
}
func (self *TBattleMoudle) OnLogin() {
}
func (self *TBattleMoudle) OnLogout() {
}

// -------------------------------------
// -- API
func Rpc_Battle_Begin(req, ack *common.NetPack, ptr interface{}) {
	battleMsg := common.NewNetPackCap(32)
	cnt := req.ReadByte()
	battleMsg.WriteByte(cnt)
	for i := byte(0); i < cnt; i++ {
		pid := req.ReadUInt32()
		if player := _FindInCache(pid); player != nil {
			battleMsg.WriteUInt32(pid)
			// pack player battle data
			battleMsg.WriteString(player.Name)

			//重新匹配战斗服
			player.Battle.loginBattleMsg.Clear()
		} else {
			ack.WriteInt8(-1)
			return
		}
	}
	battleMsg.SetRpc("rpc_cross_relay_battle_data")
	api.SendToCross(battleMsg)
	ack.WriteInt8(1)
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
		msg.Clear()
	}
	ack.WriteInt8(-1)
}
