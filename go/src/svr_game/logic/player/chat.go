/***********************************************************************
* @ 聊天模块
* @ brief
    1、别人发来的聊天信息，先缓存，作为【捎带数据】传给client
    2、client可能许久不发请求，所以client chat须在空闲期，定时探查

* @ client 探查
    1、初始20s一次，有svr回应，cd减半，最短至0.5s
    2、无回应，则倍增回初始的20s
    3、若中途有业务逻辑请求，重置探查cd

* @ author zhoumf
* @ date 2017-5-2
***********************************************************************/
package player

import (
	"common"
	"dbmgo"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	Max_Save_Chat_Msg_Cnt = 100 //缓存100条，满了删头部20条
	Del_Head_Chat_Msg_Cnt = 20
)

type TChatMoudle struct {
	PlayerID uint32 `bson:"_id"`
	ChatLst  []TChat

	owner     *TPlayer
	noSendPos int
}
type TChat struct {
	FromPid uint32
	Content string
	Time    int64
}

// -------------------------------------
// -- 框架接口
func (self *TChatMoudle) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	self.owner = player
	dbmgo.InsertSync("Chat", self)
}
func (self *TChatMoudle) WriteToDB() { dbmgo.UpdateSync("Chat", self.PlayerID, self) }
func (self *TChatMoudle) LoadFromDB(player *TPlayer) {
	dbmgo.Find("Chat", "_id", player.PlayerID, self)
	self.owner = player
}
func (self *TChatMoudle) OnLogin() {
	self.noSendPos = 0
}
func (self *TChatMoudle) OnLogout() {
	self.noSendPos = 0
}

// -------------------------------------
// -- API
func (self *TChatMoudle) CreateMsg(formPid uint32, content string) *TChat {
	ptr := &TChat{formPid, content, time.Now().Unix()}

	if len(self.ChatLst) >= Max_Save_Chat_Msg_Cnt {
		self.ChatLst = append(self.ChatLst[Del_Head_Chat_Msg_Cnt:], *ptr)
		if self.noSendPos > Del_Head_Chat_Msg_Cnt {
			self.noSendPos -= Del_Head_Chat_Msg_Cnt
		} else {
			self.noSendPos = 0
		}
		dbmgo.UpdateToDB("Chat", bson.M{"_id": self.PlayerID}, bson.M{"$set": bson.M{"chatlst": self.ChatLst}})
	} else {
		self.ChatLst = append(self.ChatLst, *ptr)
		dbmgo.UpdateToDB("Chat", bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"chatlst": ptr}})
	}
	return ptr
}
func (self *TChat) DataToBuf(buf *common.NetPack) {
	buf.WriteUInt32(self.FromPid)
	buf.WriteString(self.Content)
	buf.WriteInt64(self.Time)
}
func (self *TChat) BufToData(buf *common.NetPack) {
	self.FromPid = buf.ReadUInt32()
	self.Content = buf.ReadString()
	self.Time = buf.ReadInt64()
}
func (self *TChatMoudle) GetNoSendIdx() int {
	if len(self.ChatLst) > self.noSendPos {
		return self.noSendPos
	} else {
		return -1
	}
}
func (self *TChatMoudle) DataToBuf(buf *common.NetPack) {
	length := len(self.ChatLst)
	buf.WriteUInt16(uint16(length - self.noSendPos))
	for i := self.noSendPos; i < length; i++ {
		data := &self.ChatLst[i]
		data.DataToBuf(buf)
	}
	self.noSendPos = length
}
