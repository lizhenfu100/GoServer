package player

import (
	"common"
	"dbmgo"
	"gamelog"
)

type TFriendMoudle struct {
	PlayerID  uint32 `bson:"_id"`
	FriendLst []uint32
	ApplyLst  []uint32

	owner     *TPlayer
	inviteMsg *common.ByteBuffer
	isChange  bool
}

// -------------------------------------
// -- 框架接口
func (self *TFriendMoudle) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.InsertToDB("Friend", self)
	self.owner = player
	self._InitTempData()
}
func (self *TFriendMoudle) LoadFromDB(player *TPlayer) {
	if !dbmgo.Find("Friend", "_id", player.PlayerID, self) {
		self.InitAndInsert(player)
	}
	self.owner = player
	self._InitTempData()
}
func (self *TFriendMoudle) WriteToDB() { dbmgo.UpdateSync("Friend", self.PlayerID, self) }
func (self *TFriendMoudle) OnLogin() {
}
func (self *TFriendMoudle) OnLogout() {
}
func (self *TFriendMoudle) _InitTempData() {
	self.inviteMsg = common.NewByteBufferCap(32)
}

// -------------------------------------
//! buf
func (self *TFriendMoudle) PackFriendInfo(buf *common.NetPack) {
	ptrList := make([]*TPlayer, 0, len(self.FriendLst))
	for _, v := range self.FriendLst {
		if ptr := FindPlayerInCache(v); ptr != nil {
			ptrList = append(ptrList, ptr)
		}
	}
	buf.WriteUInt16(uint16(len(ptrList)))
	for _, v := range ptrList {
		buf.WriteUInt32(v.PlayerID)
		buf.WriteString(v.Name)
	}
}
func (self *TFriendMoudle) PackApplyInfo(buf *common.NetPack) {
	ptrList := make([]*TPlayer, 0, len(self.ApplyLst))
	for _, v := range self.ApplyLst {
		if ptr := FindPlayerInCache(v); ptr != nil {
			ptrList = append(ptrList, ptr)
		}
	}
	buf.WriteUInt8(uint8(len(ptrList)))
	for _, v := range ptrList {
		buf.WriteUInt32(v.PlayerID)
		buf.WriteString(v.Name)
	}
}

// -------------------------------------
// -- API
func (self *TFriendMoudle) RecvApply(pid uint32) int8 {
	if self.InApplyLst(pid) >= 0 {
		return -1
	}
	if self.InFriendLst(pid) >= 0 { //对方是自己好友，直接同意
		AsyncNotifyPlayer(pid, func(player *TPlayer) {
			player.Friend.AddFriend(self.PlayerID)
		})
		return 0
	}
	if self.PlayerID == pid {
		return -3
	}
	self.ApplyLst = append(self.ApplyLst, pid)
	return 0
}
func (self *TFriendMoudle) Agree(pid uint32) {
	i := self.InApplyLst(pid)
	if i < 0 {
		return
	}
	self.AddFriend(pid)
	self.ApplyLst = append(self.ApplyLst[:i], self.ApplyLst[i+1:]...)

	AsyncNotifyPlayer(pid, func(player *TPlayer) {
		player.Friend.AddFriend(self.PlayerID)
	})
}
func (self *TFriendMoudle) Refuse(pid uint32) {
	i := self.InApplyLst(pid)
	if i < 0 {
		return
	}
	self.ApplyLst = append(self.ApplyLst[:i], self.ApplyLst[i+1:]...)
}
func (self *TFriendMoudle) AddFriend(pid uint32) int8 {
	if self.InFriendLst(pid) >= 0 {
		return -2
	}
	self.FriendLst = append(self.FriendLst, pid)
	self.isChange = true
	//dbmgo.UpdateToDB("Friend", bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"friendlst": data}})
	return 0
}
func (self *TFriendMoudle) DelFriend(pid uint32) {
	if i := self.InFriendLst(pid); i >= 0 {
		//ptr := &self.FriendLst[i]
		//dbmgo.UpdateToDB("Friend", bson.M{"_id": self.PlayerID}, bson.M{"$pull": bson.M{
		//	"friendlst": bson.M{"id": ptr.ID}}})
		self.FriendLst = append(self.FriendLst[:i], self.FriendLst[i+1:]...)
		self.isChange = true
	}
}

// -------------------------------------
//! 辅助函数
func (self *TFriendMoudle) InApplyLst(pid uint32) int {
	for i := 0; i < len(self.ApplyLst); i++ {
		if pid == self.ApplyLst[i] {
			return i
		}
	}
	return -1
}
func (self *TFriendMoudle) InFriendLst(pid uint32) int {
	for i := 0; i < len(self.FriendLst); i++ {
		if pid == self.FriendLst[i] {
			return i
		}
	}
	return -1
}

// -------------------------------------
//! 加好友
func Rpc_game_friend_list(req, ack *common.NetPack, ptr interface{}) {
	player := ptr.(*TPlayer)
	player.Friend.PackFriendInfo(ack)
}
func Rpc_game_friend_apply(req, ack *common.NetPack, ptr interface{}) {
	destPid := req.ReadUInt32()
	self := ptr.(*TPlayer)

	if destPid == self.PlayerID {
		return
	}
	AsyncNotifyPlayer(destPid, func(destPtr *TPlayer) {
		destPtr.Friend.RecvApply(self.PlayerID)
	})
}
func Rpc_game_friend_agree(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*TPlayer)
	player.Friend.Agree(pid)
}
func Rpc_game_friend_refuse(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*TPlayer)
	player.Friend.Refuse(pid)
}
func Rpc_game_friend_del(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*TPlayer)
	player.Friend.DelFriend(pid)
}
func Rpc_game_search_player(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()
	player := FindPlayerInCache(pid)
	if player != nil {
		ack.WriteUInt32(player.PlayerID)
		ack.WriteString(player.Name)
	} else {
		ack.WriteUInt32(0)
	}
}

// -------------------------------------
//! 组队相关
func (self *TFriendMoudle) BeInvitedBy(p *TPlayer) {
	if self.owner.pTeam != nil { //已组队，邀请无效
		gamelog.Debug("%d is already in team", self.PlayerID)
		return
	}
	if self.inviteMsg.Size() == 0 {
		self.inviteMsg.WriteUInt32(0) //邀请的人数
	}
	self.inviteMsg.WriteUInt32(p.PlayerID)
	self.inviteMsg.WriteString(p.Name)
	cnt := self.inviteMsg.GetPos(0)
	self.inviteMsg.SetPos(0, cnt+1)
}
