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

	owner     *TPlayer
	inviteMsg *common.ByteBuffer
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
func (self *TFriendMoudle) BeInvitedBy(p *TPlayer) {
	if self.owner.pTeam != nil { //已组队，邀请无效
		fmt.Println(self.PlayerID, "is already in team", self.owner.pTeam)
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
