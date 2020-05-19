/***********************************************************************
* @ 好友系统

* @ 像微信那样，致力于将数据只保存在本地，后台只存id列表，用于关系重构
	· 牺牲及时性，免去大部分的好友状态同步，如改名、改头像
	· 玩家查看好友个人信息时，才更新；允许显示旧信息

* @ 账号下有多个角色的，如何处理？
	· 可同大区的，都算好友

* @ 跨大区社交，得做套单独的rpc，以playerId为key
	· playerId夹带路由信息

* @ 好友列表同步
	· 好友数据分为三层
		· svr_center中的accountIdList
		· svr_game中的playerIdList（aid+pid）
		· client中的friendList（aid+pid+showInfo）
	· 用svr_center数据初始化svr_game好友
		· 难以甄别svr_center多出的好友，是svr_game主动删的还是其它途径加的
	· client无好友数据，才向svr_game拉取
		· svr_game负责收集名字、头像...回给client

* @ author zhoumf
* @ date 2019-12-30
***********************************************************************/
package logic

import (
	"common"
	"common/std"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
)

const kDBTable = "Friend"

type TFriend struct { //Optimize：hash aid分表分库
	AccountId uint32 `bson:"_id"`
	Friends   std.UInt32s
}

// ------------------------------------------------------------
// - rpc
func Rpc_friend_add(req, ack *common.NetPack, _ common.Conn) {
	myId := req.ReadUInt32()
	dstId := req.ReadUInt32()
	AddFriend(myId, dstId)
	AddFriend(dstId, myId)
}
func Rpc_friend_del(req, ack *common.NetPack, _ common.Conn) {
	myId := req.ReadUInt32()
	dstId := req.ReadUInt32()
	DelFriend(myId, dstId)
	DelFriend(dstId, myId)
}
func Rpc_friend_list(req, ack *common.NetPack, _ common.Conn) {
	myId := req.ReadUInt32()
	if p := FindWithDB(myId); p != nil {
		//删除上报的，剩余即新增
		for cnt, i := req.ReadUInt16(), uint16(0); i < cnt; i++ {
			if j := p.Friends.Index(req.ReadUInt32()); j >= 0 {
				p.Friends.Del(j)
			}
		}
		//返回新好友accountId
		posInBuf, count := ack.Size(), uint16(0)
		ack.WriteUInt16(count)
		for _, aid := range p.Friends {
			ack.WriteUInt32(aid)
			count++
		}
		ack.SetUInt16(posInBuf, count)
	}
}

// ------------------------------------------------------------
// - 辅助函数
func FindWithDB(aid uint32) *TFriend {
	p := &TFriend{AccountId: aid}
	if ok, _ := dbmgo.Find(kDBTable, "_id", aid, p); ok {
		return p
	}
	return nil
}
func AddFriend(aid, dst uint32) {
	if p := FindWithDB(aid); p == nil {
		dbmgo.Insert(kDBTable, &TFriend{
			AccountId: aid,
			Friends:   []uint32{dst},
		})
	} else if i := p.Friends.Index(dst); i < 0 && dst != p.AccountId {
		p.Friends.Add(dst)
		dbmgo.UpdateId(kDBTable, p.AccountId, bson.M{"$push": bson.M{
			"friends": dst}})
	}
}
func DelFriend(aid, dst uint32) {
	if p := FindWithDB(aid); p != nil {
		if i := p.Friends.Index(dst); i >= 0 {
			p.Friends.Del(i)
			dbmgo.UpdateId(kDBTable, p.AccountId, bson.M{"$pull": bson.M{
				"friends": dst}})
		}
	}
}
