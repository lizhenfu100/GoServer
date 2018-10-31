package player

import (
	"common"
	"common/std"
	"dbmgo"
	"time"
)

const (
	Read_Mail_Delete_Time = 24 * 3600 * 7 //已读邮件多久后删除
)

type TMailModule struct {
	PlayerID  uint32 `bson:"_id"`
	MailLst   []TMail
	SvrMailId uint32 //已收到的全服邮件索引位

	/* OverWatch ECS ---- Component只含数据，System只有操作
	、按照ECS思路，TMailModule视作邮件数据，下面的函数是一个MailSystem
	、MailSystem很可能不止需要TMailModule，还会用到其它模块的东东，比如好友、网络通信模块
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
	Items   []std.IntPair
}

// -------------------------------------
// -- 框架接口
func (self *TMailModule) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.InsertToDB(kDBMail, self)
}
func (self *TMailModule) LoadFromDB(player *TPlayer) {
	if !dbmgo.Find(kDBMail, "_id", player.PlayerID, self) {
		self.InitAndInsert(player)
	}
}
func (self *TMailModule) WriteToDB() { dbmgo.UpdateIdToDB(kDBMail, self.PlayerID, self) }
func (self *TMailModule) OnLogin() {
	// 删除过期已读邮件
	timenow := time.Now().Unix()
	for i := len(self.MailLst) - 1; i >= 0; i-- { //倒过来遍历，删除就安全的
		if self.MailLst[i].IsRead == 1 && timenow >= self.MailLst[i].Time+Read_Mail_Delete_Time {
			self.MailLst = append(self.MailLst[:i], self.MailLst[i+1:]...)
		}
	}
	self.clientMailId = 0
}
func (self *TMailModule) OnLogout() {
}

// -------------------------------------
// -- API
func (self *TMailModule) CreateMail(title, from, content string, items ...std.IntPair) *TMail {
	id := dbmgo.GetNextIncId("MailId")
	pMail := &TMail{id, time.Now().Unix(), title, from, content, 0, items}
	self.MailLst = append(self.MailLst, *pMail)
	//dbmgo.UpdateToDB("Mail", bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"maillst": pMail}})
	return pMail
}
func (self *TMailModule) DelMailRead() {
	for i := len(self.MailLst) - 1; i >= 0; i-- { //倒过来遍历，删除就安全的
		if self.MailLst[i].IsRead == 1 {
			self.MailLst = append(self.MailLst[:i], self.MailLst[i+1:]...)
		}
	}
	//dbmgo.UpdateToDB("Mail", bson.M{"_id": self.PlayerID}, bson.M{"$pull": bson.M{
	//	"maillst": bson.M{"isread": 1}}})
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
	self.Items = self.Items[:0]
	for i := byte(0); i < length; i++ {
		id := buf.ReadInt()
		cnt := buf.ReadInt()
		self.Items = append(self.Items, std.IntPair{id, cnt})
	}
}
func (self *TMailModule) DataToBuf(buf *common.NetPack, pos int) {
	length := len(self.MailLst)
	buf.WriteUInt16(uint16(length - pos))
	for i := pos; i < length; i++ {
		mail := &self.MailLst[i]
		mail.DataToBuf(buf)
		self.clientMailId = mail.ID
	}
}
func (self *TMailModule) GetNoSendIdx() int {
	length := len(self.MailLst)
	for i := 0; i < length; i++ {
		if self.MailLst[i].ID > self.clientMailId {
			return i
		}
	}
	return -1
}

// -------------------------------------
//! rpc
func Rpc_game_get_mail(req, ack *common.NetPack, this *TPlayer) {
	self := this.Mail
	if pos := self.GetNoSendIdx(); pos >= 0 {
		ack.WriteInt8(1)
		self.DataToBuf(ack, pos)
	} else {
		ack.WriteInt8(-1)
	}
}
func Rpc_game_read_mail(req, ack *common.NetPack, this *TPlayer) {
	id := req.ReadUInt32()

	self := this.Mail
	for i := 0; i < len(self.MailLst); i++ {
		mail := &self.MailLst[i]
		if mail.ID == id {
			if len(mail.Items) > 0 {
				//TODO: 给奖励
			}
			mail.IsRead = 1
			return
		}
	}
}
func Rpc_game_del_mail(req, ack *common.NetPack, this *TPlayer) {
	id := req.ReadUInt32()

	self := this.Mail
	for i := len(self.MailLst) - 1; i >= 0; i-- { //倒过来遍历，删除就安全的
		mail := &self.MailLst[i]
		if mail.ID == id {
			if mail.IsRead == 0 && len(mail.Items) > 0 {
				//TODO: 给奖励
			}
			self.MailLst = append(self.MailLst[:i], self.MailLst[i+1:]...)
			return
		}
	}
}
func Rpc_game_take_all_mail_item(req, ack *common.NetPack, this *TPlayer) {
	self := this.Mail
	//self.Mail.CreateMail(0, "测试", "zhoumf", "content")
	for i := len(self.MailLst) - 1; i >= 0; i-- { //倒过来遍历，删除就安全的
		mail := &self.MailLst[i]
		if len(mail.Items) > 0 {
			//TODO: 给奖励
			self.MailLst = append(self.MailLst[:i], self.MailLst[i+1:]...)
		}
	}
	self.DataToBuf(ack, 0)
}
