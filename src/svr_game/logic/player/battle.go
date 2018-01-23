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
	"generate_out/rpc/enum"
	"netConfig"
	"svr_game/api"
	"tcp"
)

type TBattleMoudle struct {
	PlayerID uint32 `bson:"_id"`

	loginBattleMsg *common.ByteBuffer
	isShowWaitUI   bool
}

// -------------------------------------
// -- 框架接口
func (self *TBattleMoudle) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.InsertToDB("Battle", self)
	self._InitTempData()
}
func (self *TBattleMoudle) LoadFromDB(player *TPlayer) {
	if !dbmgo.Find("Battle", "_id", player.PlayerID, self) {
		self.InitAndInsert(player)
	}
	self._InitTempData()
}
func (self *TBattleMoudle) WriteToDB() { dbmgo.UpdateSync("Battle", self.PlayerID, self) }
func (self *TBattleMoudle) OnLogin() {
}
func (self *TBattleMoudle) OnLogout() {
}
func (self *TBattleMoudle) _InitTempData() {
	self.loginBattleMsg = common.NewByteBufferCap(32)
}

// -------------------------------------
// -- Rpc
func Rpc_game_battle_begin(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	if self.pTeam == nil || self.pTeam.lst[0] != self {
		return
	}
	api.CallRpcCross(enum.Rpc_cross_relay_battle_data, func(buf *common.NetPack) {
		buf.WriteString(netConfig.G_Local_Meta.Version)
		buf.WriteByte(byte(len(self.pTeam.lst)))
		for _, ptr := range self.pTeam.lst {
			buf.WriteUInt32(ptr.PlayerID)
			// pack player battle data
			buf.WriteString(ptr.Name)

			//重新匹配战斗服
			ptr.Battle.loginBattleMsg.Clear()

			// 通知队员，开等待界面
			if ptr != self {
				ptr.AsyncNotify(func(p *TPlayer) {
					p.Battle.isShowWaitUI = true
				})
			}
		}
	}, nil)
}
func Rpc_game_battle_ack(req, ack *common.NetPack, conn *tcp.TCPConn) {
	battleSvrOutIP := req.ReadString()
	battleSvrPort := req.ReadUInt16()
	cnt := req.ReadByte()
	for i := byte(0); i < cnt; i++ {
		pid := req.ReadUInt32()
		//通知client登录战斗服
		AsyncNotifyPlayer(pid, func(player *TPlayer) {
			buf := player.Battle.loginBattleMsg
			buf.WriteString(battleSvrOutIP)
			buf.WriteUInt16(battleSvrPort)
		})
	}
}
func Rpc_game_probe_login_battle(req, ack *common.NetPack, ptr interface{}) {
	player := ptr.(*TPlayer)

	if player.Battle.loginBattleMsg.Size() > 0 {
		ack.WriteInt8(1)
		ack.WriteBuf(player.Battle.loginBattleMsg.Data())
		player.Battle.loginBattleMsg.Clear()
	} else {
		ack.WriteInt8(-1)
	}
}
