package player

import (
	"common"
	"dbmgo"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

type TFriendMoudle struct {
	PlayerID  uint32 `bson:"_id"`
	FriendLst []TFriend
	ApplyLst  []TFriend

	owner        *TPlayer
	inviteMsg    *common.ByteBuffer
	pTeamInfo    *[]*TPlayerBase //同一队伍的人引用同一地址，但仅队长能操作数据，别人只读
	isTeamChange bool
}
type TFriend struct {
	ID   uint32
	Name string
}

// -------------------------------------
// -- 框架接口
func (self *TFriendMoudle) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	self.owner = player
	dbmgo.InsertSync("Friend", self)

	self._InitTempData()
}
func (self *TFriendMoudle) WriteToDB() { dbmgo.UpdateSync("Friend", self.PlayerID, self) }
func (self *TFriendMoudle) LoadFromDB(player *TPlayer) {
	dbmgo.Find("Friend", "_id", player.PlayerID, self)
	self.owner = player

	self._InitTempData()
}
func (self *TFriendMoudle) OnLogin() {
}
func (self *TFriendMoudle) OnLogout() {
	self.ExitTeam()
}
func (self *TFriendMoudle) _InitTempData() {
	self.inviteMsg = common.NewByteBufferCap(32)
}

// -------------------------------------
//! buf
func (self *TFriend) DataToBuf(buf *common.NetPack) {
	buf.WriteUInt32(self.ID)
	buf.WriteString(self.Name)
}
func (self *TFriend) BufToData(buf *common.NetPack) {
	self.ID = buf.ReadUInt32()
	self.Name = buf.ReadString()
}
func (self *TFriendMoudle) DataToBuf(buf *common.NetPack) {
	length := len(self.FriendLst)
	buf.WriteUInt16(uint16(length))
	for i := 0; i < length; i++ {
		self.FriendLst[i].DataToBuf(buf)
	}
	length = len(self.ApplyLst)
	buf.WriteUInt16(uint16(length))
	for i := 0; i < length; i++ {
		self.ApplyLst[i].DataToBuf(buf)
	}
}

// -------------------------------------
// -- API
func (self *TFriendMoudle) FindFriend(pid uint32) *TFriend {
	length := len(self.FriendLst)
	for i := 0; i < length; i++ {
		friend := &self.FriendLst[i]
		if friend.ID == pid {
			return friend
		}
	}
	return nil
}
func (self *TFriendMoudle) RecvApply(pid uint32, name string) int8 {
	if self.InApplyLst(pid) >= 0 {
		return -1
	}
	if self.InFriendLst(pid) >= 0 {
		return -2
	}
	if self.PlayerID == pid {
		return -3
	}
	data := TFriend{pid, name}
	self.ApplyLst = append(self.ApplyLst, data)
	dbmgo.UpdateToDB("Friend", bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"applylst": data}})
	return 0
}
func (self *TFriendMoudle) Agree(pid uint32) {
	i := self.InApplyLst(pid)
	if i < 0 {
		return
	}
	ptr := &self.ApplyLst[i]

	if self.AddFriend(ptr.ID, ptr.Name) >= 0 {
		dbmgo.UpdateToDB("Friend", bson.M{"_id": self.PlayerID}, bson.M{"$pull": bson.M{
			"applylst": bson.M{"id": ptr.ID}}})
	}
	self.ApplyLst = append(self.ApplyLst[:i], self.ApplyLst[i+1:]...)

	AsyncNotifyPlayer(ptr.ID, func(player *TPlayer) {
		player.Friend.AddFriend(self.PlayerID, self.owner.Name)
	})
}
func (self *TFriendMoudle) Refuse(pid uint32) {
	i := self.InApplyLst(pid)
	if i < 0 {
		return
	}
	ptr := &self.ApplyLst[i]
	dbmgo.UpdateToDB("Friend", bson.M{"_id": self.PlayerID}, bson.M{"$pull": bson.M{
		"applylst": bson.M{"id": ptr.ID}}})
	self.ApplyLst = append(self.ApplyLst[:i], self.ApplyLst[i+1:]...)
}
func (self *TFriendMoudle) AddFriend(pid uint32, name string) int8 {
	if self.InFriendLst(pid) >= 0 {
		return -2
	}
	data := TFriend{pid, name}
	self.FriendLst = append(self.FriendLst, data)
	dbmgo.UpdateToDB("Friend", bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"friendlst": data}})
	return 0
}
func (self *TFriendMoudle) DelFriend(pid uint32) {
	if i := self.InFriendLst(pid); i >= 0 {
		ptr := &self.FriendLst[i]
		dbmgo.UpdateToDB("Friend", bson.M{"_id": self.PlayerID}, bson.M{"$pull": bson.M{
			"friendlst": bson.M{"id": ptr.ID}}})
		self.FriendLst = append(self.FriendLst[:i], self.FriendLst[i+1:]...)
	}
}

// -------------------------------------
//! 辅助函数
func (self *TFriendMoudle) InApplyLst(pid uint32) int {
	for i := 0; i < len(self.ApplyLst); i++ {
		if pid == self.ApplyLst[i].ID {
			return i
		}
	}
	return -1
}
func (self *TFriendMoudle) InFriendLst(pid uint32) int {
	for i := 0; i < len(self.FriendLst); i++ {
		if pid == self.FriendLst[i].ID {
			return i
		}
	}
	return -1
}

// -------------------------------------
//! 加好友
func Rpc_Friend_List(req, ack *common.NetPack, ptr interface{}) {
	player := ptr.(*TPlayer)
	player.Friend.DataToBuf(ack)
}
func Rpc_Friend_Apply(req, ack *common.NetPack, ptr interface{}) {
	destPid := req.ReadUInt32()
	self := ptr.(*TPlayer)

	AsyncNotifyPlayer(destPid, func(destPtr *TPlayer) {
		destPtr.Friend.RecvApply(self.PlayerID, self.Name)
	})
}
func Rpc_Friend_Agree(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*TPlayer)
	player.Friend.Agree(pid)
}
func Rpc_Friend_Refuse(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*TPlayer)
	player.Friend.Refuse(pid)
}
func Rpc_Friend_Del(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*TPlayer)
	player.Friend.DelFriend(pid)
}

// -------------------------------------
//! 组队相关
func Rpc_Create_Team(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)

	lst := make([]*TPlayerBase, 0, 5)
	lst = append(lst, &self.TPlayerBase)
	self.Friend.pTeamInfo = &lst
}
func Rpc_Exit_Team(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	self.Friend.ExitTeam()
}
func Rpc_Get_Team_Info(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	fmt.Println("Team_Info", self.Friend.pTeamInfo)
	if self.Friend.pTeamInfo == nil {
		ack.WriteByte(0)
	} else {
		ack.WriteByte(byte(len(*self.Friend.pTeamInfo)))
		for _, p := range *self.Friend.pTeamInfo {
			ack.WriteUInt32(p.PlayerID)
			ack.WriteString(p.Name)
		}
	}
}
func Rpc_Invite_Friend(req, ack *common.NetPack, ptr interface{}) { //邀请别人
	destPid := req.ReadUInt32()
	self := ptr.(*TPlayer)

	AsyncNotifyPlayer(destPid, func(dest *TPlayer) {
		dest.Friend.BeInvitedBy(&self.TPlayerBase)
	})
}
func Rpc_Agree_Join_Team(req, ack *common.NetPack, ptr interface{}) { //同意加队
	self := ptr.(*TPlayer)

	if self.Friend.pTeamInfo != nil {
		ack.WriteByte(0)
		return
	}
	destPid := req.ReadUInt32()
	if dest := _FindInCache(destPid); dest != nil && dest.Friend.pTeamInfo != nil {
		self.Friend.pTeamInfo = dest.Friend.pTeamInfo //! readonly

		fmt.Println("Join_Team", self.Friend.pTeamInfo)
		// 下发队伍信息
		team := *self.Friend.pTeamInfo
		ack.WriteByte(byte(len(team)))
		for _, p := range team {
			ack.WriteUInt32(p.PlayerID)
			ack.WriteString(p.Name)
		}

		// 通知队长，加自己
		dest.AsyncNotify(func(p *TPlayer) {
			p.Friend.JoinToMyTeam(&self.TPlayerBase)
		})
	} else {
		ack.WriteByte(0)
	}
}
func (self *TFriendMoudle) BeInvitedBy(pBase *TPlayerBase) {
	if self.pTeamInfo != nil { //已组队，邀请无效
		fmt.Println(self.PlayerID, "is already in team", self.pTeamInfo)
		return
	}
	if self.inviteMsg.Size() == 0 {
		self.inviteMsg.WriteUInt32(0) //邀请的人数
	}
	self.inviteMsg.WriteUInt32(pBase.PlayerID)
	self.inviteMsg.WriteString(pBase.Name)
	cnt := self.inviteMsg.GetPos(0)
	self.inviteMsg.SetPos(0, cnt+1)
}
func (self *TFriendMoudle) JoinToMyTeam(pBase *TPlayerBase) {
	if self.pTeamInfo == nil {
		return
	}
	fmt.Println("JoinToMyTeam", self.pTeamInfo)
	// 广播给其它队友
	for _, v := range *self.pTeamInfo {
		AsyncNotifyPlayer(v.PlayerID, func(dest *TPlayer) {
			dest.Friend.isTeamChange = true
		})
	}
	*self.pTeamInfo = append(*self.pTeamInfo, pBase)
}
func (self *TFriendMoudle) _ExitFromMyTeam(destPid uint32) {
	if self.pTeamInfo == nil {
		return
	}
	team := self.pTeamInfo
	for i := 0; i < len(*team); i++ {
		pid := (*team)[i].PlayerID
		if pid == destPid {
			*team = append((*team)[:i], (*team)[i+1:]...)
			i--
		} else {
			// 广播给其它队友
			AsyncNotifyPlayer(pid, func(ptr *TPlayer) {
				ptr.Friend.isTeamChange = true
			})
		}
	}
}
func (self *TFriendMoudle) ExitTeam() {
	if self.pTeamInfo == nil {
		return
	}
	fmt.Println("ExitTeam", self.pTeamInfo)

	captainPid := (*self.pTeamInfo)[0].PlayerID
	if self.PlayerID == captainPid {
		self._ExitFromMyTeam(self.PlayerID)
	} else {
		AsyncNotifyPlayer(captainPid, func(dest *TPlayer) {
			dest.Friend._ExitFromMyTeam(self.PlayerID)
		})
	}
	self.pTeamInfo = nil
}
