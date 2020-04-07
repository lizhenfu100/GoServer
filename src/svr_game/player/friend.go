package player

import (
	"common"
	"common/std"
	"dbmgo"
	"generate_out/rpc/enum"
	"gopkg.in/mgo.v2/bson"
	"netConfig"
	"netConfig/meta"
	"svr_game/conf"
)

const kDBFriend = "friend"

type TFriendModule struct {
	Pid     uint64 `bson:"_id"`
	Friends std.UInt64s
}

// ------------------------------------------------------------
// -- 框架接口
func (self *TFriendModule) InitAndInsert(p *TPlayer) {
	self.Pid = common.PidNew(p.AccountID, conf.Const.LoginSvrId, meta.G_Local.SvrID)
	dbmgo.Insert(kDBFriend, self)
}
func (self *TFriendModule) LoadFromDB(p *TPlayer) {
	if ok, _ := dbmgo.Find(kDBFriend, "_id", p.PlayerID, self); !ok {
		self.InitAndInsert(p)
	}
}
func (self *TFriendModule) WriteToDB() { dbmgo.UpdateId(kDBFriend, self.Pid, self) }
func (self *TFriendModule) OnLogin() {
	myId := common.PidToAid(self.Pid)
	for _, pid := range self.Friends {
		netConfig.CallRpcGateway(common.PidToAid(pid), enum.Rpc_client_friend_login, func(buf *common.NetPack) {
			buf.WriteUInt32(myId)
		}, nil)
	}
}
func (self *TFriendModule) OnLogout() {
	myId := common.PidToAid(self.Pid)
	for _, pid := range self.Friends {
		netConfig.CallRpcGateway(common.PidToAid(pid), enum.Rpc_client_friend_logout, func(buf *common.NetPack) {
			buf.WriteUInt32(myId)
		}, nil)
	}
}
func (self *TFriendModule) InitFriends() {
	if p, ok := netConfig.GetRpcRand("friend"); ok { //公司好友数据，初始化游戏好友
		p.CallRpc(enum.Rpc_friend_list, func(buf *common.NetPack) {
			buf.WriteUInt32(common.PidToAid(self.Pid))
			buf.WriteUInt16(uint16(len(self.Friends)))
			for _, v := range self.Friends {
				buf.WriteUInt32(common.PidToAid(v))
			}
		}, func(recvBuf *common.NetPack) {
			for cnt, i := recvBuf.ReadUInt16(), uint16(0); i < cnt; i++ {
				aid := recvBuf.ReadUInt32()
				ptr := &TPlayerBase{} //本区aid下的角色，加好友
				if ok, _ := dbmgo.Find(kDBPlayer, "accountid", aid, ptr); ok {
					pid := common.PidNew(ptr.AccountID, conf.Const.LoginSvrId, meta.G_Local.SvrID)
					self.Friends.Add(pid)
				}
			}
		})
	}
}

// ------------------------------------------------------------
// --
func Rpc_game_friend_add(req, ack *common.NetPack, this *TPlayer) {
	dstId := req.ReadUInt64()

	this.friend.AddFriend(dstId)
	other := &TFriendModule{Pid: dstId}
	if ok, _ := dbmgo.Find(kDBFriend, "_id", dstId, other); ok {
		other.AddFriend(this.friend.Pid)
	} else {
		other.Friends.Add(this.friend.Pid)
		dbmgo.Insert(kDBFriend, other)
	}
}
func Rpc_game_friend_del(req, ack *common.NetPack, this *TPlayer) {
	dstId := req.ReadUInt64()

	this.friend.DelFriend(dstId)
	other := &TFriendModule{Pid: dstId}
	if ok, _ := dbmgo.Find(kDBFriend, "_id", dstId, other); ok {
		other.DelFriend(this.friend.Pid)
	}
}
func Rpc_game_friend_list(req, ack *common.NetPack, this *TPlayer) {
	list := make(std.UInt64s, len(this.friend.Friends))
	copy(list, this.friend.Friends)
	//删除上报的，剩余即新增
	for cnt, i := req.ReadUInt16(), uint16(0); i < cnt; i++ {
		if j := list.Index(req.ReadUInt64()); j >= 0 {
			list.Del(j)
		}
	}
	//返回新好友showInfo
	posInBuf, count := ack.Size(), uint16(0)
	ack.WriteUInt16(count)
	var base TPlayerBase
	for _, pid := range list {
		//Optimize:先收集本大区(同个db)，其它大区的走proxy转发
		if ok, _ := dbmgo.Find(kDBPlayer, "_id", pid, &base); ok {
			base.GetShowInfo().DataToBuf(ack)
			count++
		}
	}
	ack.SetUInt16(posInBuf, count)
}

// ------------------------------------------------------------
// - 辅助函数
func (self *TFriendModule) AddFriend(dst uint64) {
	if i := self.Friends.Index(dst); i < 0 && dst != self.Pid {
		self.Friends.Add(dst)
		dbmgo.UpdateId(kDBFriend, self.Pid, bson.M{"$push": bson.M{
			"friends": dst}})
	}
}
func (self *TFriendModule) DelFriend(dst uint64) {
	if i := self.Friends.Index(dst); i >= 0 {
		self.Friends.Del(i)
		dbmgo.UpdateId(kDBFriend, self.Pid, bson.M{"$pull": bson.M{
			"friends": dst}})
	}
}
