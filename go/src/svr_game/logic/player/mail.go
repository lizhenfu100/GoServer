package player

import (
	"common"
	"dbmgo"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type TMailMoudle struct {
	PlayerID     uint32 `bson:"_id"`
	MailLst      []TMail
	SvrMailId    uint32 //已收到的全服邮件索引位，登录时收后续的
	owner        *TPlayer
	clientMailId uint32 //已发给client的
}
type TMail struct {
	ID      uint32 `bson:"_id"`
	Time    int64
	Title   string
	From    string
	Content string
	IsRead  byte
	Items   []common.IntPair
}

func (self *TMailMoudle) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	self.owner = player
	dbmgo.InsertSync("Mail", self)
}
func (self *TMailMoudle) WriteToDB() { dbmgo.UpdateSync("Mail", self.PlayerID, self) }
func (self *TMailMoudle) LoadFromDB(player *TPlayer) {
	dbmgo.Find("Mail", "_id", player.PlayerID, self)
	self.owner = player
}
func (self *TMailMoudle) OnLogin() {
	SendSvrMailOnLogin(self.owner)
}
func (self *TMailMoudle) OnLogout() {
	self.clientMailId = 0
}

func (self *TMailMoudle) CreateMail(title, from, content string, items ...common.IntPair) *TMail {
	id := dbmgo.GetNextIncId("MailId")
	pMail := &TMail{id, time.Now().Unix(), title, from, content, 0, items}
	self.MailLst = append(self.MailLst, *pMail)
	dbmgo.UpdateToDB("Mail", bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"maillst": pMail}})
	return pMail
}
func (self *TMailMoudle) DelMail(id uint32) {
	for i := 0; i < len(self.MailLst); i++ {
		if self.MailLst[i].ID == id {
			self.MailLst = append(self.MailLst[:i], self.MailLst[i+1:]...)
			dbmgo.UpdateToDB("Mail", bson.M{"_id": self.PlayerID}, bson.M{"$pull": bson.M{
				"maillst": bson.M{"id": id}}})
		}
	}
}
func (self *TMailMoudle) DelMailRead() {
	for i := 0; i < len(self.MailLst); i++ {
		if self.MailLst[i].IsRead == 1 {
			self.MailLst = append(self.MailLst[:i], self.MailLst[i+1:]...)
		}
	}
	dbmgo.UpdateToDB("Mail", bson.M{"_id": self.PlayerID}, bson.M{"$pull": bson.M{
		"maillst": bson.M{"isread": 1}}})
}

// buf := common.NewNetPack()
// MailToBuf(&buf.ByteBuffer)
func (self *TMail) MailToBuf(buf *common.ByteBuffer) {
	buf.WriteUInt32(self.ID)
	buf.WriteInt64(self.Time)
	buf.WriteString(self.Title)
	buf.WriteString(self.From)
	buf.WriteString(self.Content)
	buf.WriteByte(self.IsRead)
	length := len(self.Items)
	buf.WriteByte(byte(length))
	for i := 0; i < length; i++ {
		item := &self.Items[i]
		buf.WriteInt(item.ID)
		buf.WriteInt(item.Cnt)
	}
}
func (self *TMail) BufToMail(buf *common.ByteBuffer) {
	self.ID = buf.ReadUInt32()
	self.Time = buf.ReadInt64()
	self.Title = buf.ReadString()
	self.From = buf.ReadString()
	self.Content = buf.ReadString()
	self.IsRead = buf.ReadByte()
	length := buf.ReadByte()
	for i := byte(0); i < length; i++ {
		id := buf.ReadInt()
		cnt := buf.ReadInt()
		self.Items = append(self.Items, common.IntPair{id, cnt})
	}
}
func (self *TMailMoudle) GetNoSendMailIdx() int {
	length := len(self.MailLst)
	for i := 0; i < length; i++ {
		if self.MailLst[i].ID > self.clientMailId {
			return i
		}
	}
	return -1
}
func (self *TMailMoudle) MailLstToBuf(buf *common.ByteBuffer, pos int) {
	length := len(self.MailLst)
	buf.WriteUInt32(uint32(length - pos))
	for i := pos; i < length; i++ {
		mail := &self.MailLst[i]
		mail.MailToBuf(buf)
		self.clientMailId = mail.ID
	}
}
