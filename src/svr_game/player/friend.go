package player

import (
	"common"
	"dbmgo"
	"generate_out/rpc/enum"
	"gopkg.in/mgo.v2/bson"
	"netConfig"
	"netConfig/meta"
	"svr_game/conf"
)

const kDBFriend = "friend"

type TFriendModule struct {
	Uid     common.Uid `bson:"_id"`
	Friends []common.Uid
}

// ------------------------------------------------------------
// -- 框架接口
func (self *TFriendModule) InitAndInsert(p *TPlayer) {
	self.Uid = common.UidNew(p.AccountID, conf.Const.LoginSvrId, meta.G_Local.SvrID)
	dbmgo.Insert(kDBFriend, self)
}
func (self *TFriendModule) LoadFromDB(p *TPlayer) {
	if ok, _ := dbmgo.Find(kDBFriend, "_id", p.PlayerID, self); !ok {
		self.InitAndInsert(p)
	}
}
func (self *TFriendModule) WriteToDB() { dbmgo.UpdateId(kDBFriend, self.Uid, self) }
func (self *TFriendModule) OnLogin() {
	myId := self.Uid.ToAid()
	for _, v := range self.Friends {
		netConfig.CallRpcGateway(v.ToAid(),
			enum.Rpc_client_friend_login, func(buf *common.NetPack) {
				buf.WriteUInt32(myId)
			}, nil)
	}
}
func (self *TFriendModule) OnLogout() {
	myId := self.Uid.ToAid()
	for _, v := range self.Friends {
		netConfig.CallRpcGateway(v.ToAid(),
			enum.Rpc_client_friend_logout, func(buf *common.NetPack) {
				buf.WriteUInt32(myId)
			}, nil)
	}
}
func (self *TFriendModule) InitFriends() {
	if p, ok := netConfig.GetRpcRand("friend"); ok { //公司好友数据，初始化游戏好友
		p.CallRpc(enum.Rpc_friend_list, func(buf *common.NetPack) {
			buf.WriteUInt32(self.Uid.ToAid())
			buf.WriteUInt16(uint16(len(self.Friends)))
			for _, v := range self.Friends {
				buf.WriteUInt32(v.ToAid())
			}
		}, func(recvBuf *common.NetPack) {
			ptr := &TPlayerBase{}
			for cnt, i := recvBuf.ReadUInt16(), uint16(0); i < cnt; i++ {
				aid := recvBuf.ReadUInt32() //本区aid下的角色，加好友
				if ok, _ := dbmgo.Find(kDBPlayer, "accountid", aid, ptr); ok {
					uid := common.UidNew(aid, conf.Const.LoginSvrId, meta.G_Local.SvrID)
					self.Friends = append(self.Friends, uid)
				}
			}
		})
	}
}

// ------------------------------------------------------------
// --
func Rpc_game_friend_add(req, ack *common.NetPack, this *TPlayer) {
	dstId := common.Uid(req.ReadUInt64())

	this.friend.AddFriend(dstId)
	other := &TFriendModule{Uid: dstId}
	if ok, _ := dbmgo.Find(kDBFriend, "_id", dstId, other); ok {
		other.AddFriend(this.friend.Uid)
	} else {
		other.Friends = append(other.Friends, this.friend.Uid)
		dbmgo.Insert(kDBFriend, other)
	}
}
func Rpc_game_friend_del(req, ack *common.NetPack, this *TPlayer) {
	dstId := common.Uid(req.ReadUInt64())

	this.friend.DelFriend(dstId)
	other := &TFriendModule{Uid: dstId}
	if ok, _ := dbmgo.Find(kDBFriend, "_id", dstId, other); ok {
		other.DelFriend(this.friend.Uid)
	}
}
func Rpc_game_friend_list(req, ack *common.NetPack, this *TPlayer) {
	list := make([]common.Uid, len(this.friend.Friends))
	copy(list, this.friend.Friends)
	//删除上报的，剩余即新增
	for cnt, i := req.ReadUInt16(), uint16(0); i < cnt; i++ {
		for j, v := 0, req.ReadUInt64(); j < len(list); j++ {
			if v == uint64(list[j]) {
				list = append(list[:j], list[j+1:]...)
				break
			}
		}
	}
	//返回新好友showInfo
	posInBuf, count, base := ack.Size(), uint16(0), TPlayerBase{}
	ack.WriteUInt16(count)
	for i := 0; i < len(list); i++ {
		//Optimize:先收集本大区(同个db)，其它大区的走proxy转发
		if ok, _ := dbmgo.Find(kDBPlayer, "_id", list[i], &base); ok {
			base.GetShowInfo().DataToBuf(ack)
			count++
		}
	}
	ack.SetUInt16(posInBuf, count)
}

// ------------------------------------------------------------
// - 辅助函数
func (self *TFriendModule) AddFriend(dst common.Uid) {
	if dst == self.Uid {
		return
	}
	for i := 0; i < len(self.Friends); i++ {
		if dst == self.Friends[i] {
			return
		}
	}
	self.Friends = append(self.Friends, dst)
	dbmgo.UpdateId(kDBFriend, self.Uid, bson.M{"$push": bson.M{
		"friends": dst}})
}
func (self *TFriendModule) DelFriend(dst common.Uid) {
	for i := 0; i < len(self.Friends); i++ {
		if dst == self.Friends[i] {
			self.Friends = append(self.Friends[:i], self.Friends[i+1:]...)
			dbmgo.UpdateId(kDBFriend, self.Uid, bson.M{"$pull": bson.M{
				"friends": dst}})
		}
	}
}
