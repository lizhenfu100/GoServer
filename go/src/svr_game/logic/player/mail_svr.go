package player

import (
	"common"
	"dbmgo"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var (
	g_svr_mail   []TMail
	g_mail_mutex sync.RWMutex
)

func InitSvrMailLst() {
	//只读一个月内的
	dbmgo.FindAll("MailSvr", bson.M{"time": bson.M{"$gt": time.Now().Unix() - 30*24*3600}}, &g_svr_mail)
}
func CreateSvrMail(title, from, content string, items ...common.IntPair) {
	id := dbmgo.GetNextIncId("SvrMailId")
	pMail := &TMail{id, 0, time.Now().Unix(), title, from, content, 0, items}
	dbmgo.InsertToDB("MailSvr", pMail)

	g_mail_mutex.Lock()
	g_svr_mail = append(g_svr_mail, *pMail)
	g_mail_mutex.Unlock()

	// 遍历全服，太不可控了
	// ForEachOnlinePlayer(func(player *TPlayer) {
	// 	player.Mail.SendSvrMail(pMail)
	// })
}
func (self *TMailMoudle) SendSvrMail(pData *TMail) bool {
	if pData.ID <= self.SvrMailId {
		return false // received
	}
	self.SvrMailId = pData.ID
	self.MailLst = append(self.MailLst, *pData)
	mail := &self.MailLst[len(self.MailLst)-1]
	mail.ID = dbmgo.GetNextIncId("MailId")
	dbmgo.UpdateToDB("Mail", bson.M{"_id": self.PlayerID}, bson.M{"$push": bson.M{"maillst": mail}})
	// self.owner.WriteMsg( mail.MailToBuf() )
	return true
}
func (self *TMailMoudle) SendSvrMailAll() {
	g_mail_mutex.RLock()
	defer g_mail_mutex.RUnlock()
	length := len(g_svr_mail)
	for i := length - 1; i >= 0; i-- {
		if false == self.SendSvrMail(&g_svr_mail[i]) {
			return
		}
	}
}
