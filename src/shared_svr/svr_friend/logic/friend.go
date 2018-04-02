package logic

import (
	"common"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"tcp"
)

type TFriendModule struct {
	AccountId uint32 `bson:"_id"`
	FriendIDs []uint32
}

var g_friend_cache sync.Map //make(map[uint32]*TFriendModule, 5000)

func FindWithDB(aid uint32) *TFriendModule {
	if v, ok := g_friend_cache.Load(aid); ok {
		return v.(*TFriendModule)
	} else {
		ptr := new(TFriendModule)
		if dbmgo.Find("Friend", "_id", aid, ptr) {
			AddCache(ptr)
			return ptr
		}
	}
	return nil
}
func AddCache(ptr *TFriendModule) { g_friend_cache.Store(ptr.AccountId, ptr) }
func DelCache(ptr *TFriendModule) { g_friend_cache.Delete(ptr.AccountId) }

// ------------------------------------------------------------
// -- rpc
func Rpc_friend_get_friend_list(req, ack *common.NetPack) {
	aid := req.ReadUInt32()

	if self := FindWithDB(aid); self != nil {
		ack.WriteByte(byte(len(self.FriendIDs)))
		for _, id := range self.FriendIDs {
			ack.WriteUInt32(id)
		}
	}
}
func Rpc_friend_add(req, ack *common.NetPack, conn *tcp.TCPConn) {
	aid := req.ReadUInt32()

	if self := FindWithDB(aid); self != nil {
		if i := self.InFriendLst(aid); i < 0 {
			self.FriendIDs = append(self.FriendIDs, aid)
			dbmgo.UpdateToDB("Friend", bson.M{"_id": self.AccountId}, bson.M{"$push": bson.M{
				"friendids": aid}})
		}
	}
}
func Rpc_friend_del(req, ack *common.NetPack, conn *tcp.TCPConn) {
	aid := req.ReadUInt32()

	if self := FindWithDB(aid); self != nil {
		if i := self.InFriendLst(aid); i >= 0 {
			self.FriendIDs = append(self.FriendIDs[:i], self.FriendIDs[i+1:]...)
			dbmgo.UpdateToDB("Friend", bson.M{"_id": self.AccountId}, bson.M{"$pull": bson.M{
				"friendids": aid}})
		}
	}
}

// ------------------------------------------------------------
// -- 辅助函数
func (self *TFriendModule) InFriendLst(aid uint32) int {
	for i := 0; i < len(self.FriendIDs); i++ {
		if aid == self.FriendIDs[i] {
			return i
		}
	}
	return -1
}
