package account

import (
	"dbmgo"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type TAccountMgr struct {
	sync.RWMutex
	NameToPtr map[string]*TAccount
	IdToPtr   map[uint32]*TAccount
}

var G_AccountMgr TAccountMgr

func (self *TAccountMgr) Init() {
	self.IdToPtr = make(map[uint32]*TAccount, 5000)
	self.NameToPtr = make(map[string]*TAccount, 5000)
	//只载入一个月内登录过的
	var accountLst []TAccount
	dbmgo.FindAll("Account", bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - 30*24*3600}}, &accountLst)
	for i := 0; i < len(accountLst); i++ {
		self.AddToCache(&accountLst[i])
	}
	println("load active account form db: ", len(accountLst))
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
		self.AddToCache(account)
		return account
	}
	return nil
}
func (self *TAccountMgr) GetAccountByName(name string) *TAccount {
	self.RLock()
	ret := self.NameToPtr[name]
	self.RUnlock()
	if ret == nil {
		account := new(TAccount)
		if ok := dbmgo.Find("Account", "name", name, account); ok {
			self.AddToCache(account)
			return account
		}
	}
	return ret
}
func (self *TAccountMgr) GetAccountById(accountId uint32) *TAccount {
	self.RLock()
	ret := self.IdToPtr[accountId]
	self.RUnlock()
	if ret == nil {
		account := new(TAccount)
		if ok := dbmgo.Find("Account", "_id", accountId, account); ok {
			self.AddToCache(account)
			return account
		}
	}
	return ret
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

// -------------------------------------
//! 辅助函数
func (self *TAccountMgr) AddToCache(account *TAccount) {
	self.Lock()
	self.IdToPtr[account.AccountID] = account
	self.NameToPtr[account.Name] = account
	self.Unlock()
}
func (self *TAccountMgr) DelToCache(account *TAccount) {
	self.Lock()
	delete(self.IdToPtr, account.AccountID)
	delete(self.NameToPtr, account.Name)
	self.Unlock()
}
