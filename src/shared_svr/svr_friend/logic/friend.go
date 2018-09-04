package logic

import (
	"common"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
)

const kDBTable = "Friend"

type TFriendModule struct {
	AccountId uint32 `bson:"_id"`
	FriendIDs []uint32
}

// ------------------------------------------------------------
// - rpc
func Rpc_friend_add(req, ack *common.NetPack, this *TFriendModule) {
	aid := req.ReadUInt32()

	if i := this.InFriendLst(aid); i < 0 {
		this.FriendIDs = append(this.FriendIDs, aid)
		dbmgo.UpdateToDB(kDBTable, bson.M{"_id": this.AccountId}, bson.M{"$push": bson.M{
			"friendids": aid}})
	}
}
func Rpc_friend_del(req, ack *common.NetPack, this *TFriendModule) {
	aid := req.ReadUInt32()

	if i := this.InFriendLst(aid); i >= 0 {
		this.FriendIDs = append(this.FriendIDs[:i], this.FriendIDs[i+1:]...)
		dbmgo.UpdateToDB(kDBTable, bson.M{"_id": this.AccountId}, bson.M{"$pull": bson.M{
			"friendids": aid}})
	}
}

// ------------------------------------------------------------
// - 业务节点提取好友列表
func Rpc_friend_get_friend_list(req, ack *common.NetPack) {
	aid := req.ReadUInt32()

	if self := FindWithDB(aid); self != nil {
		ack.WriteByte(byte(len(self.FriendIDs)))
		for _, id := range self.FriendIDs {
			ack.WriteUInt32(id)
		}
	}
}

// ------------------------------------------------------------
// - 辅助函数
func FindWithDB(aid uint32) *TFriendModule {
	ptr := new(TFriendModule)
	if !dbmgo.Find(kDBTable, "_id", aid, ptr) {
		dbmgo.InsertToDB(kDBTable, ptr)
	}
	return ptr
}
func (self *TFriendModule) InFriendLst(aid uint32) int {
	for i := 0; i < len(self.FriendIDs); i++ {
		if aid == self.FriendIDs[i] {
			return i
		}
	}
	return -1
}
