package player

import (
	"common/std"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	g_svr_mail   []TMail
	g_mail_mutex sync.RWMutex
)

func InitSvrMailDB() {
	//只读一个月内的
	dbmgo.FindAll(kDBMailSvr, bson.M{"time": bson.M{"$gt": time.Now().Unix() - 30*24*3600}}, &g_svr_mail)
}
func CreateSvrMail(title, from, content string, items ...std.IntPair) {
	id := dbmgo.GetNextIncId("SvrMailId")
	pMail := &TMail{id, time.Now().Unix(), title, from, content, 0, items}
	dbmgo.InsertToDB(kDBMailSvr, pMail)

	g_mail_mutex.Lock()
	g_svr_mail = append(g_svr_mail, *pMail)
	g_mail_mutex.Unlock()
}
func (self *TMailModule) SendSvrMail(mail *TMail) bool {
	if mail.ID <= self.SvrMailId {
		return false
	}
	self.SvrMailId = mail.ID
	mail.ID = dbmgo.GetNextIncId("MailId")
	dbmgo.UpdateToDB(kDBMail, bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"maillst": mail}})
	self.MailLst = append(self.MailLst, *mail)
	// self.owner.WriteMsg( mail.MailToBuf() )
	return true
}
func (self *TMailModule) SendSvrMailAll() {
	g_mail_mutex.RLock()
	defer g_mail_mutex.RUnlock()
	length := len(g_svr_mail)
	for i := length - 1; i >= 0; i-- {
		if false == self.SendSvrMail(&g_svr_mail[i]) {
			return
		}
	}
}
