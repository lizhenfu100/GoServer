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
	mutex         sync.Mutex
	autoAccountID uint32
	NameToId      map[string]uint32
	IdToPtr       map[uint32]*TAccount
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

	dbmgo.Find_Desc("Account", "_id", 1, &accountLst)
	if len(accountLst) > 0 {
		self.autoAccountID = accountLst[0].AccountID + 1
	} else {
		self.autoAccountID = Account_ID_Begin
	}
}
func (self *TAccountMgr) AddNewAccount(name, password string) *TAccount {
	if _, ok := self.NameToId[name]; ok {
		return nil
	}
	account := &TAccount{
		AccountID:  self._GetNextAccountID(),
		Name:       name,
		Password:   password,
		CreateTime: time.Now().Unix(),
	}
	if err := dbmgo.InsertSync("Account", account); err != nil {
		self._AddToCache(account)
		return account
	}
	return nil
}
func (self *TAccountMgr) GetAccountByName(name string) *TAccount {
	if id, ok := self.NameToId[name]; ok {
		return self.IdToPtr[id]
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
func (self *TAccountMgr) _GetNextAccountID() (ret uint32) {
	self.mutex.Lock()
	ret = self.autoAccountID
	self.autoAccountID++
	self.mutex.Unlock()
	return
}
func (self *TAccountMgr) _AddToCache(account *TAccount) {
	self.mutex.Lock()
	self.IdToPtr[account.AccountID] = account
	self.NameToId[account.Name] = account.AccountID
	self.mutex.Unlock()
}
func CreateLoginToken() string {
	return "chillyroom"
}
