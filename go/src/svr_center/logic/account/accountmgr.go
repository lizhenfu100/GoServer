package account

import (
	"dbmgo"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	Account_ID_Begin  = 10000          //10000以下的ID留给服务器使用
	Login_Active_Time = 30 * 24 * 3600 //一个月内登录过的，算活跃玩家
)

type TAccountMgr struct {
	sync.RWMutex
	NameToId map[string]uint32
	IdToPtr  map[uint32]*TAccount
}

var G_AccountMgr TAccountMgr

func (self *TAccountMgr) Init() {
	self.IdToPtr = make(map[uint32]*TAccount, 5000)
	self.NameToId = make(map[string]uint32, 5000)
	//只载入活跃玩家
	var accountLst []TAccount
	dbmgo.FindAll("Account", bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - Login_Active_Time}}, &accountLst)
	for i := 0; i < len(accountLst); i++ {
		self._AddToCache(&accountLst[i])
	}
}
func (self *TAccountMgr) AddNewAccount(name, password string) *TAccount {
	account := &TAccount{
		Name:       name,
		Password:   password,
		CreateTime: time.Now().Unix(),
	}
	if dbmgo.Find("Account", "name", name, account) {
		return nil
	}
	account.AccountID = dbmgo.GetNextIncId("AccountId")

	if dbmgo.InsertSync("Account", account) {
		self._AddToCache(account)
		return account
	}
	return nil
}
func (self *TAccountMgr) GetAccountByName(name string) *TAccount {
	self.RLock()
	id := self.NameToId[name]
	self.RUnlock()
	if id > 0 {
		self.RLock()
		account := self.IdToPtr[id]
		self.RUnlock()
		return account
	} else {
		account := new(TAccount)
		if ok := dbmgo.Find("Account", "name", name, account); ok {
			self._AddToCache(account)
			return account
		}
	}
	return nil
}
func (self *TAccountMgr) ResetPassword(name, password, newpassword string) bool {
	if account := self.GetAccountByName(name); account != nil {
		if account.Password == password {
			account.Password = newpassword
			dbmgo.UpdateToDB("Account", bson.M{"_id": account.AccountID}, bson.M{"$set": bson.M{
				"password": newpassword}})
			return true
		}
	}
	return false
}

//! 辅助函数
func (self *TAccountMgr) _AddToCache(account *TAccount) {
	self.Lock()
	self.IdToPtr[account.AccountID] = account
	self.NameToId[account.Name] = account.AccountID
	self.Unlock()
}
func CreateLoginToken() string {
	return "chillyroom"
}
