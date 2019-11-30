package logic

import (
	"common"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"nets/tcp"
)

const kDBTable = "Friend"

type TFriendModule struct {
	AccountId uint32 `bson:"_id"`
	FriendIDs []uint32
	//TODO：双向好友系统
}

// ------------------------------------------------------------
// - rpc
func Rpc_friend_add(req, ack *common.NetPack, this *TFriendModule) {
	aid := req.ReadUInt32()

	if i := this.InFriends(aid); i < 0 && aid != this.AccountId {
		this.FriendIDs = append(this.FriendIDs, aid)
		dbmgo.UpdateId(kDBTable, this.AccountId, bson.M{"$push": bson.M{
			"friendids": aid}})
	}
}
func Rpc_friend_del(req, ack *common.NetPack, this *TFriendModule) {
	aid := req.ReadUInt32()

	if i := this.InFriends(aid); i >= 0 {
		this.FriendIDs = append(this.FriendIDs[:i], this.FriendIDs[i+1:]...)
		dbmgo.UpdateId(kDBTable, this.AccountId, bson.M{"$pull": bson.M{
			"friendids": aid}})
	}
}
func Rpc_friend_get_friend_list(req, ack *common.NetPack, conn *tcp.TCPConn) {
	aid := req.ReadUInt32()

	if self := FindWithDB(aid); self != nil {
		ack.WriteUInt16(uint16(len(self.FriendIDs)))
		for _, id := range self.FriendIDs {
			ack.WriteUInt32(id)
		}
	}
}

// ------------------------------------------------------------
// - 辅助函数
func FindWithDB(aid uint32) *TFriendModule {
	ptr := &TFriendModule{AccountId: aid}
	if ok, _ := dbmgo.Find(kDBTable, "_id", aid, ptr); !ok {
		dbmgo.Insert(kDBTable, ptr)
	}
	return ptr
}
func (self *TFriendModule) InFriends(aid uint32) int {
	for i := 0; i < len(self.FriendIDs); i++ {
		if aid == self.FriendIDs[i] {
			return i
		}
	}
	return -1
}
