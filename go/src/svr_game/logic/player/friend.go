package player

import (
	"common"
	"dbmgo"

	"gopkg.in/mgo.v2/bson"
)

type TFriendMoudle struct {
	PlayerID  uint32 `bson:"_id"`
	FriendLst []TFriend
	ApplyLst  []TFriend

	owner *TPlayer
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
}
func (self *TFriendMoudle) WriteToDB() { dbmgo.UpdateSync("Friend", self.PlayerID, self) }
func (self *TFriendMoudle) LoadFromDB(player *TPlayer) {
	dbmgo.Find("Friend", "_id", player.PlayerID, self)
	self.owner = player
}
func (self *TFriendMoudle) OnLogin() {
}
func (self *TFriendMoudle) OnLogout() {
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
func (self *TFriendMoudle) Agree(i byte) {
	if i >= byte(len(self.ApplyLst)) {
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
func (self *TFriendMoudle) Refuse(i byte) {
	if i >= byte(len(self.ApplyLst)) {
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
//! rpc
func Rpc_Friend_List(req, ack *common.NetPack, ptr interface{}) {
	player := ptr.(*TPlayer)
	player.Friend.DataToBuf(ack)
}
func Rpc_Friend_Apply(req, ack *common.NetPack, ptr interface{}) {
	destPid := req.ReadUInt32()

	AsyncNotifyPlayer(destPid, func(destPtr *TPlayer) {
		player := ptr.(*TPlayer)
		destPtr.Friend.RecvApply(player.PlayerID, player.Name)
	})
}
func Rpc_Friend_Agree(req, ack *common.NetPack, ptr interface{}) {
	pos := req.ReadByte()

	player := ptr.(*TPlayer)
	player.Friend.Agree(pos)
}
func Rpc_Friend_Refuse(req, ack *common.NetPack, ptr interface{}) {
	pos := req.ReadByte()

	player := ptr.(*TPlayer)
	player.Friend.Refuse(pos)
}
func Rpc_Friend_Del(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*TPlayer)
	player.Friend.DelFriend(pid)
}
