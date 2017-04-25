package mail

import (
	"common"
	"dbmgo"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type TMailMoudle struct {
	PlayerID uint32 `bson:"_id"`
	MailLst  []TMail
}
type TMail struct {
	ID      uint32
	Time    int64
	Title   string
	From    string
	Content string
	IsRead  byte
	Items   []common.IntPair
}

func (self *TMailMoudle) InitAndInsert(id uint32) {
	self.PlayerID = id
	dbmgo.InsertSync("Mail", self)
}
func (self *TMailMoudle) WriteToDB()           { dbmgo.UpdateSync("Mail", self.PlayerID, self) }
func (self *TMailMoudle) LoadFromDB(id uint32) { dbmgo.Find("Mail", "_id", id, self) }
func (self *TMailMoudle) OnLogin()             {}
func (self *TMailMoudle) OnLogout()            {}

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
func (self *TMail) MailToBuf(buf *common.NetPack) {
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
func (self *TMail) BufToMail(buf *common.NetPack) {
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
