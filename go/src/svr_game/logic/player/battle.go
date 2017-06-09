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
}
func (self *TBattleMoudle) WriteToDB() { dbmgo.UpdateSync("Battle", self.PlayerID, self) }
func (self *TBattleMoudle) LoadFromDB(player *TPlayer) {
	dbmgo.Find("Battle", "_id", player.PlayerID, self)
	self.owner = player
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
		if player := _FindPlayerInCache(pid); player != nil {
			// pack player battle data
			battleMsg.WriteUInt32(pid)
		} else {
			ack.WriteInt8(-1)
			return
		}
	}
	battleMsg.SetRpc("rpc_relay_battle_data")
	api.SendToCross(battleMsg)
	ack.WriteInt8(1)
}
func Rpc_Battle_Ack(conn *tcp.TCPConn, msg *common.NetPack) {
	battleSvrIP := msg.ReadString()
	battleSvrPort := msg.ReadUInt16()
	cnt := msg.ReadByte()
	for i := byte(0); i < cnt; i++ {
		pid := msg.ReadUInt32()
		idx := msg.ReadUInt32() //战斗服分配的玩家内存索引
		//通知client登录战斗服
		AsyncNotifyPlayer(pid, func(player *TPlayer) {
			buf := common.NewByteBufferCap(32)
			buf.WriteString(battleSvrIP)
			buf.WriteUInt16(battleSvrPort)
			buf.WriteUInt32(idx)
			player.Battle.loginBattleMsg = buf
		})
	}
}
func Rpc_Probe_Login_Battle(req, ack *common.NetPack, ptr interface{}) {
	player := ptr.(*TPlayer)

	if player.Battle.loginBattleMsg != nil {
		ack.WriteInt8(1)
		ack.WriteBuf(player.Battle.loginBattleMsg.DataPtr)

		player.Battle.loginBattleMsg = nil
	}
	ack.WriteInt8(-1)
}
