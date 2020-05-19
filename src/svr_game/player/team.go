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
	、组队服，数量不多，可否逐个同步取数据(缓存"取到数据的节点信息"，下回直接拿该节点，无需遍历了)……避免逻辑层维护缓存匹配关系

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
	"nets/tcp"
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
func (self *Team) Clear()        { self.list = self.list[:0] }
func (self *Team) IsEmpty() bool { return len(self.list) == 0 }
func (self *Team) Have(aid uint32) bool {
	for i := 0; i < len(self.list); i++ {
		if self.list[i].show.AccountId == aid {
			return true
		}
	}
	return false
}

func (self *Team) DataToBuf(buf *common.NetPack) {
	buf.WriteByte(byte(len(self.list)))
	for i := 0; i < len(self.list); i++ {
		self.list[i].DataToBuf(buf)
	}
}
func (self *Team) BufToData(buf *common.NetPack) {
	var v TBattleModule
	self.list = self.list[:0]
	for cnt, i := buf.ReadByte(), byte(0); i < cnt; i++ {
		v.BufToData(buf)
		self.list = append(self.list, v)
	}
}
func (self *Team) ShowInfoToBuf(buf *common.NetPack) {
	buf.WriteByte(byte(len(self.list)))
	for i := 0; i < len(self.list); i++ {
		self.list[i].show.DataToBuf(buf)
	}
}

// ------------------------------------------------------------
// -- Rpc
func Rpc_game_battle_begin(req, ack *common.NetPack, this *TPlayer) {
	if this.team.IsEmpty() {
		this.team.list = append(this.team.list, this.battle)
	}
	gameMode := req.ReadUInt8()

	CallRpcCross(enum.Rpc_cross_relay_battle_data, func(buf *common.NetPack) {
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
		CallRpcPlayer(ptr.show.AccountId, enum.Rpc_client_show_wait_ui, func(buf *common.NetPack) {
		}, nil)
	}
}

var g_cache_cross = make(map[int]*tcp.TCPConn)

func CallRpcCross(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if p := meta.GetByRand("cross"); p != nil {
		if conn := GetCrossConn(p.SvrID); conn != nil {
			conn.CallEx(rid, sendFun, recvFun)
			return
		}
	}
	gamelog.Error("cross nil: rpcId(%d)", rid)
}
func GetCrossConn(svrId int) *tcp.TCPConn {
	conn, _ := g_cache_cross[svrId]
	if conn == nil || conn.IsClose() {
		conn = netConfig.GetTcpConn("cross", svrId)
		g_cache_cross[svrId] = conn
	}
	return conn
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
	} else {
		gamelog.Error("Aid(%d) is in team", this.AccountID)
	}
}
func Rpc_game_become_captain(req, ack *common.NetPack, this *TPlayer) {
	this.team.BufToData(req)
}
func Rpc_game_get_team_show_info(req, ack *common.NetPack, this *TPlayer) {
	gamelog.Debug("Team_Info: %v", this.team)
	this.team.ShowInfoToBuf(ack)
}
func Rpc_game_team_chat(req, ack *common.NetPack, this *TPlayer) {
	srcId := req.ReadUInt32()
	srcName := req.ReadString()
	content := req.ReadString()

	for i := 0; i < len(this.team.list); i++ {
		if p := &this.team.list[i]; p.show.AccountId != srcId {
			CallRpcPlayer(p.show.AccountId, enum.Rpc_client_team_chat, func(buf *common.NetPack) {
				buf.WriteString(srcName)
				buf.WriteString(content)
			}, nil)
		}
	}
}

// ------------------------------------------------------------
// 玩家组队
func (self *TPlayer) JoinMyTeam(dest *TBattleModule) {
	gamelog.Debug("JoinMyTeam: %v", self.team)

	if self.team.Have(dest.show.AccountId) {
		return
	}
	// 队内广播加入事件 -- 后台队伍数据，仅队长持有即可，同步到队员客户端，队员后台不必有
	for i := 0; i < len(self.team.list); i++ {
		ptr := &self.team.list[i]
		CallRpcPlayer(ptr.show.AccountId, enum.Rpc_client_on_other_join_team, func(buf *common.NetPack) {
			dest.show.DataToBuf(buf)
		}, nil)
	}
	self.team.list = append(self.team.list, *dest)
	CallRpcPlayer(dest.show.AccountId, enum.Rpc_client_on_self_join_team, func(buf *common.NetPack) {
		self.team.ShowInfoToBuf(buf)
		gamelog.Debug("%v", self.team)
	}, nil)
}
func (self *TPlayer) ExitMyTeam(accountId uint32) {
	gamelog.Debug("ExitMyTeam: %v", self.team)

	for i := len(self.team.list) - 1; i >= 0; i-- { //倒序遍历，删除更安全的
		if p := self.team.list[i]; p.show.AccountId == accountId {
			self.team.list = append(self.team.list[:i], self.team.list[i+1:]...)
		} else {
			// 队内广播给离开事件
			CallRpcPlayer(p.show.AccountId, enum.Rpc_client_on_exit_team, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
			}, nil)
		}
	}
	if self.AccountID == accountId { //自己离队，转移数据
		if len(self.team.list) > 0 {
			CallRpcPlayer(self.team.list[0].show.AccountId, enum.Rpc_game_become_captain, func(buf *common.NetPack) {
				self.team.DataToBuf(buf)
			}, nil)
			self.team.Clear()
		}
	}
}
