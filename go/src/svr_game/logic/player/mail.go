package player

import (
	"common"
	"dbmgo"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	Read_Mail_Delete_Time = 24 * 3600 * 7
)

type TMailMoudle struct {
	PlayerID  uint32 `bson:"_id"`
	MailLst   []TMail
	SvrMailId uint32 //已收到的全服邮件索引位

	/* OverWatch ECS ---- Component只含数据，System只有操作
	、按照ECS思路，TMailMoudle视作邮件数据，下面的函数是一个MailSystem
	、MailSystem很可能不止需要TMailMoudle，还会用到其它模块的东东，比如好友、网络通信模块
	、所以加了 owner *TPlayer ，需要时去调用其它模块
	、代码上看，就引入了一整结构，破坏封装 (每个模块都有可能改别人关心的数据，造成影响传递)
	、ECS将数据组织成 Component ，各个 System 自己声明要关注的 Component ，还能指定 readonly ，简洁干净多了
	*/
	// owner        *TPlayer
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

// -------------------------------------
// -- 框架接口
func (self *TMailMoudle) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.InsertToDB("Mail", self)
}
func (self *TMailMoudle) LoadFromDB(player *TPlayer) {
	if !dbmgo.Find("Mail", "_id", player.PlayerID, self) {
		self.InitAndInsert(player)
	}
}
func (self *TMailMoudle) WriteToDB() { dbmgo.UpdateSync("Mail", self.PlayerID, self) }
func (self *TMailMoudle) OnLogin() {
	// 删除过期已读邮件
	timenow := time.Now().Unix()
	for i := 0; i < len(self.MailLst); i++ {
		if self.MailLst[i].IsRead == 1 && timenow >= self.MailLst[i].Time+Read_Mail_Delete_Time {
			self.MailLst = append(self.MailLst[:i], self.MailLst[i+1:]...)
			i--
		}
	}
	self.clientMailId = 0
}
func (self *TMailMoudle) OnLogout() {
}

// -------------------------------------
// -- API
func (self *TMailMoudle) CreateMail(title, from, content string, items ...common.IntPair) *TMail {
	id := dbmgo.GetNextIncId("MailId")
	pMail := &TMail{id, time.Now().Unix(), title, from, content, 0, items}
	self.MailLst = append(self.MailLst, *pMail)
	dbmgo.UpdateToDB("Mail", bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"maillst": pMail}})
	return pMail
}
func (self *TMailMoudle) DelMail(id uint32) {
	for i := 0; i < len(self.MailLst); i++ {
		mail := &self.MailLst[i]
		if mail.ID == id {
			if mail.IsRead == 0 && len(mail.Items) > 0 {
				//TODO: 给奖励
			}
			self.MailLst = append(self.MailLst[:i], self.MailLst[i+1:]...)
			dbmgo.UpdateToDB("Mail", bson.M{"_id": self.PlayerID}, bson.M{"$pull": bson.M{
				"maillst": bson.M{"id": id}}})
			return
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

// -------------------------------------
//! buf
func (self *TMail) DataToBuf(buf *common.NetPack) {
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
func (self *TMail) BufToData(buf *common.NetPack) {
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
func (self *TMailMoudle) GetNoSendIdx() int {
	length := len(self.MailLst)
	for i := 0; i < length; i++ {
		if self.MailLst[i].ID > self.clientMailId {
			return i
		}
	}
	return -1
}
func (self *TMailMoudle) DataToBuf(buf *common.NetPack, pos int) {
	length := len(self.MailLst)
	buf.WriteUInt16(uint16(length - pos))
	for i := pos; i < length; i++ {
		mail := &self.MailLst[i]
		mail.DataToBuf(buf)
		self.clientMailId = mail.ID
	}
}

// -------------------------------------
//! rpc
func Rpc_game_get_mail(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	pos := 0 //pos := self.Mail.GetNoSendIdx()
	if pos >= 0 {
		ack.WriteInt8(1)
		self.Mail.DataToBuf(ack, pos)
	} else {
		ack.WriteInt8(-1)
	}
}
func Rpc_game_read_mail(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	id := req.ReadUInt32()
	for i := 0; i < len(self.Mail.MailLst); i++ {
		mail := &self.Mail.MailLst[i]
		if mail.ID == id && len(mail.Items) == 0 {
			mail.IsRead = 1
			return
		}
	}
}
func Rpc_game_del_mail(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	id := req.ReadUInt32()
	self.Mail.DelMail(id)
}
func Rpc_game_take_mail_item(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	id := req.ReadUInt32()
	for i := 0; i < len(self.Mail.MailLst); i++ {
		mail := &self.Mail.MailLst[i]
		if mail.ID == id && len(mail.Items) > 0 {
			mail.IsRead = 1
			//TODO: 给奖励
			return
		}
	}
}
func Rpc_game_take_all_mail_item(req, ack *common.NetPack, ptr interface{}) {
	self := ptr.(*TPlayer)
	//self.Mail.CreateMail(0, "测试", "zhoumf", "content")
	for i := 0; i < len(self.Mail.MailLst); i++ {
		mail := &self.Mail.MailLst[i]
		if len(mail.Items) > 0 {
			//TODO: 给奖励
			self.Mail.MailLst = append(self.Mail.MailLst[:i], self.Mail.MailLst[i+1:]...)
			i--
		}
	}
	self.Mail.DataToBuf(ack, 0)
}
