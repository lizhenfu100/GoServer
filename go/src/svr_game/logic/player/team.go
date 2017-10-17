package player

import (
	"common"
	"fmt"
)

//同一队伍的人引用同一地址，但仅队长能操作数据，别人只读
type TeamData struct {
	lst      []*TPlayer
	isChange bool
	chatLst  []TeamChat
	chatPos  map[uint32]int //要发给pid的索引位
}
type TeamChat struct {
	pid  uint32
	name string
	str  string
}

// -------------------------------------
//! 组队相关
func Rpc_game_create_team(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)

	if self.pTeam != nil {
		self.ExitTeam()
	}
	self.pTeam = &TeamData{
		lst:     []*TPlayer{self},
		chatPos: make(map[uint32]int),
	}
}
func Rpc_game_exit_team(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	self.ExitTeam()
}
func Rpc_game_get_team_info(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	fmt.Println("Team_Info", self.pTeam)
	if self.pTeam == nil {
		ack.WriteByte(0)
	} else {
		ack.WriteByte(byte(len(self.pTeam.lst)))
		for _, p := range self.pTeam.lst {
			ack.WriteUInt32(p.PlayerID)
			ack.WriteString(p.Name)
		}
	}
}
func Rpc_game_invite_friend(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	if self.pTeam == nil {
		return
	}
	destPid := req.ReadUInt32()
	AsyncNotifyPlayer(destPid, func(dest *TPlayer) {
		dest.Friend.BeInvitedBy(self)
	})
}
func Rpc_game_agree_join_team(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	if self.pTeam != nil {
		return
	}
	destPid := req.ReadUInt32()
	if captain := FindPlayerInCache(destPid); captain != nil && captain.pTeam != nil { //! readonly

		fmt.Println("Agree_Join_Team", captain.pTeam)

		// 通知队长，加自己
		captain.AsyncNotify(func(p *TPlayer) {
			p.JoinToMyTeam(self)
		})
	}
}
func (self *TPlayer) JoinToMyTeam(dest *TPlayer) {
	fmt.Println("JoinToMyTeam", self.pTeam)
	if self.pTeam == nil || dest.pTeam != nil {
		return
	}
	self.pTeam.lst = append(self.pTeam.lst, dest)
	dest.pTeam = self.pTeam

	for _, v := range self.pTeam.lst {
		v.AsyncNotify(func(p *TPlayer) {
			if p.pTeam != nil {
				p.pTeam.isChange = true
			}
		})
	}
}
func (self *TPlayer) _ExitFromMyTeam(destPid uint32) {
	if self.pTeam == nil {
		return
	}
	for i := len(self.pTeam.lst) - 1; i >= 0; i-- { //倒过来遍历，删除就安全的
		ptr := self.pTeam.lst[i]
		if ptr.PlayerID == destPid {
			self.pTeam.lst = append(self.pTeam.lst[:i], self.pTeam.lst[i+1:]...)
		} else {
			ptr.AsyncNotify(func(p *TPlayer) { // 广播给其它队友
				if p.pTeam != nil {
					p.pTeam.isChange = true
				}
			})
		}
	}
}
func (self *TPlayer) ExitTeam() {
	fmt.Println("ExitTeam", self.pTeam)
	if self.pTeam == nil {
		return
	}
	captain := self.pTeam.lst[0]
	if self.PlayerID == captain.PlayerID {
		self._ExitFromMyTeam(self.PlayerID)
	} else {
		captain.AsyncNotify(func(p *TPlayer) {
			p._ExitFromMyTeam(self.PlayerID)
		})
	}
	self.pTeam = nil
}

// -------------------------------------
//! 聊天
func Rpc_game_send_team_chat(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	pid := self.PlayerID
	if self.pTeam == nil {
		return
	}
	str := req.ReadString()
	self.pTeam.chatLst = append(self.pTeam.chatLst, TeamChat{pid, self.Name, str})

	if pos := self.pTeam.GetNoSendIdx(pid); pos >= 0 {
		self.pTeam.DataToBuf(ack, pid)
	}
}
func (self *TeamData) GetNoSendIdx(pid uint32) int {
	pos := self.chatPos[pid]
	length := len(self.chatLst)
	if length > pos {
		return pos
	} else {
		return -1
	}
}
func (self *TeamData) DataToBuf(buf *common.NetPack, pid uint32) {
	//下发从pos起始的内容
	pos := self.chatPos[pid]
	length := len(self.chatLst)
	buf.WriteUInt16(uint16(length - pos))
	for i := pos; i < length; i++ {
		v := &self.chatLst[i]
		buf.WriteUInt32(v.pid)
		buf.WriteString(v.name)
		buf.WriteString(v.str)
	}
	self.chatPos[pid] = length
}
