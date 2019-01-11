/***********************************************************************
* @ 组队交互
* @ brief
* @ 与游戏服逻辑放一块
	、后台仅队长持有队伍数据，其它玩家仅客户端有即可
	、换队长时，须将队伍数据发到新队长节点
	、队员所在客户端，须能Call到队长节点 -- 要求Gateway互连

* @ 抽离为独立的交互服
	、用redis或http服，存队伍数据，大家都到同个子节点交互
	、这样做，组队逻辑简便，但需维护“玩家-组队服”匹配关系
	、组队服，数量不多，可否同步逐个取数据(缓存取到数据的节点信息，下回直接拿该节点，无需遍历了)……避免逻辑层维护缓存匹配关系

* @ author zhoumf
* @ date 2018-3-26
***********************************************************************/
package player

import (
	"common"
	"gamelog"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
)

type Team struct {
	list []TBattleModule
}

// ------------------------------------------------------------
// -- 框架接口
func (self *Team) InitAndInsert(*TPlayer) {}
func (self *Team) LoadFromDB(*TPlayer)    {}
func (self *Team) WriteToDB()             {}
func (self *Team) OnLogin()               {}
func (self *Team) OnLogout()              {}

// ------------------------------------------------------------
// -- 队伍数据
func (self *Team) Clear()             { self.list = self.list[:0] }
func (self *Team) IsEmpty() bool      { return len(self.list) == 0 }
func (self *TPlayer) IsCaptain() bool { return self.team.list[0].aid == self.AccountID }

func (self *Team) DataToBuf(buf *common.NetPack) {
	buf.WriteByte(byte(len(self.list)))
	for i := 0; i < len(self.list); i++ {
		ptr := &self.list[i]
		ptr.DataToBuf(buf)
	}
}
func (self *Team) BufToData(buf *common.NetPack) {
	var v TBattleModule
	length := buf.ReadByte()
	self.list = self.list[:0]
	for i := byte(0); i < length; i++ {
		v.BufToData(buf)
		self.list = append(self.list, v)
	}
}

// ------------------------------------------------------------
// -- Rpc
func Rpc_game_battle_begin(req, ack *common.NetPack, this *TPlayer) {
	if this.team.IsEmpty() {
		this.team.list = append(this.team.list, this.battle)
	}
	gameMode := req.ReadUInt8()

	netConfig.CallRpcCross(enum.Rpc_cross_relay_battle_data, func(buf *common.NetPack) {
		buf.WriteString(meta.G_Local.Version)
		buf.WriteUInt8(gameMode)
		this.team.DataToBuf(buf)
	}, func(recvBuf *common.NetPack) {
		recvBuf.ReadUInt32() //战斗服人数
		battleSvrOutIP := recvBuf.ReadString()
		battleSvrPort := recvBuf.ReadUInt16()
		cnt := recvBuf.ReadByte()
		for i := byte(0); i < cnt; i++ {
			accountId := recvBuf.ReadUInt32()

			// 通知client登录战斗服
			CallRpcPlayer(accountId, enum.Rpc_client_login_battle, func(buf *common.NetPack) {
				buf.WriteString(battleSvrOutIP)
				buf.WriteUInt16(battleSvrPort)
			}, nil)
		}
	})

	// 通知队员，开等待界面
	for i := 0; i < len(this.team.list); i++ {
		ptr := &this.team.list[i]
		CallRpcPlayer(ptr.aid, enum.Rpc_client_show_wait_ui, func(buf *common.NetPack) {}, nil)
	}
}
func Rpc_game_exit_team(req, ack *common.NetPack, this *TPlayer) {
	accountId := req.ReadUInt32()

	this.ExitMyTeam(accountId)
}
func Rpc_game_join_team(req, ack *common.NetPack, this *TPlayer) {
	var v TBattleModule
	v.BufToData(req)

	if this.team.IsEmpty() {
		this.team.list = append(this.team.list, this.battle)
	}
	this.JoinMyTeam(&v)
}
func Rpc_game_agree_invite(req, ack *common.NetPack, this *TPlayer) {
	captainId := req.ReadUInt32()

	if len(this.team.list) <= 1 {
		CallRpcPlayer(captainId, enum.Rpc_game_join_team, func(buf *common.NetPack) {
			this.battle.DataToBuf(buf)
		}, nil)
	}
}
func Rpc_game_become_captain(req, ack *common.NetPack, this *TPlayer) {
	this.team.BufToData(req)
}
func Rpc_game_get_team_info(req, ack *common.NetPack, this *TPlayer) {
	gamelog.Debug("Team_Info: %v", this.team)
	this.team.DataToBuf(ack)
}
func Rpc_game_team_chat(req, ack *common.NetPack, this *TPlayer) {
	srcId := req.ReadUInt32()
	content := req.ReadString()

	srcName := ""
	for i := 0; i < len(this.team.list); i++ {
		ptr := &this.team.list[i]
		if ptr.aid == srcId {
			srcName = ptr.name
		}
	}
	for i := 0; i < len(this.team.list); i++ {
		ptr := &this.team.list[i]
		if ptr.aid != srcId {
			CallRpcPlayer(ptr.aid, enum.Rpc_client_team_chat, func(buf *common.NetPack) {
				buf.WriteString(srcName)
				buf.WriteString(content)
			}, nil)
		}
	}
}

// ------------------------------------------------------------
// 玩家组队
func (self *TPlayer) JoinMyTeam(dest *TBattleModule) {
	gamelog.Debug("JoinTeam: %v", self.team)

	// 队内广播加入事件 -- 后台队伍数据，仅队长持有即可，同步到队员客户端，队员后台不必有
	for i := 0; i < len(self.team.list); i++ {
		ptr := &self.team.list[i]
		CallRpcPlayer(ptr.aid, enum.Rpc_client_on_other_join_team, func(buf *common.NetPack) {
			dest.DataToBuf(buf)
		}, nil)
	}

	self.team.list = append(self.team.list, *dest)
	CallRpcPlayer(dest.aid, enum.Rpc_client_on_self_join_team, func(buf *common.NetPack) {
		self.team.DataToBuf(buf)
	}, nil)
}
func (self *TPlayer) ExitMyTeam(accountId uint32) {
	gamelog.Debug("ExitTeam: %v", self.team)

	for i := len(self.team.list) - 1; i >= 0; i-- { //倒序遍历，删除更安全的
		ptr := self.team.list[i]
		if ptr.aid == accountId {
			self.team.list = append(self.team.list[:i], self.team.list[i+1:]...)
		} else {
			// 队内广播给离开事件
			CallRpcPlayer(ptr.aid, enum.Rpc_client_on_exit_team, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
			}, nil)
		}
	}
	if self.AccountID == accountId { //自己离队，转移数据
		if len(self.team.list) > 0 {
			CallRpcPlayer(self.team.list[0].aid, enum.Rpc_game_become_captain, func(buf *common.NetPack) {
				self.team.DataToBuf(buf)
			}, nil)
			self.team.Clear()
		}
	}
}
