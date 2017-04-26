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
	g_mail_mutex sync.Mutex
)

func InitSvrMailLst() {
	//只读一个月内的
	dbmgo.FindAll("MailSvr", bson.M{"time": bson.M{"$gt": time.Now().Unix() - 30*24*3600}}, &g_svr_mail)
}
func CreateSvrMail(title, from, content string, items ...common.IntPair) {
	id := dbmgo.GetNextIncId("SvrMailId")
	pMail := &TMail{id, time.Now().Unix(), title, from, content, 0, items}
	dbmgo.InsertToDB("MailSvr", pMail)

	g_mail_mutex.Lock()
	g_svr_mail = append(g_svr_mail, *pMail)
	g_mail_mutex.Unlock()

	ForEachOnlinePlayer(func(player *TPlayer) {
		SendSvrMail(player, pMail)
	})
}
func SendSvrMail(player *TPlayer, pData *TMail) {
	player.Mail.SvrMailId = pData.ID
	player.Mail.MailLst = append(player.Mail.MailLst, *pData)
	mail := &player.Mail.MailLst[len(player.Mail.MailLst)-1]
	mail.ID = dbmgo.GetNextIncId("MailId")
	dbmgo.UpdateToDB("Mail", bson.M{"_id": player.PlayerID}, bson.M{"$push": bson.M{"maillst": mail}})
	// player.WriteMsg( mail.MailToBuf() )
}
func SendSvrMailOnLogin(player *TPlayer) {
	g_mail_mutex.Lock()
	defer g_mail_mutex.Unlock()
	length := len(g_svr_mail)
	for i := 0; i < length; i++ {
		SendSvrMail(player, &g_svr_mail[i])
	}
}
