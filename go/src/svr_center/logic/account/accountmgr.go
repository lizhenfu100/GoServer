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

var G_AccountMgr TAccountMgr

type TAccount struct {
	AccountID  uint32 `bson:"_id"` //账号ID
	Name       string //账户名
	Password   string //密码
	CreateTime int64
	LoginTime  int64
	LoginCount uint32
	LoginSvrID uint32
	Forbidden  bool //是否禁用
}
type TAccountMgr struct {
	mutex         sync.Mutex
	autoAccountID uint32
	NameToId      map[string]uint32
	IdToPtr       map[uint32]*TAccount
}

func (self *TAccountMgr) Init() {
	self.IdToPtr = make(map[uint32]*TAccount, 1024)
	self.NameToId = make(map[string]uint32, 1024)

	//只载入活跃玩家
	var accountLst []TAccount
	dbmgo.FindAll("Account", bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - Login_Active_Time}}, &accountLst)
	for i := 0; i < len(accountLst); i++ {
		self._AddToCache(&accountLst[i])
	}

	dbmgo.Find_Desc("Account", "_id", 1, &accountLst)
	self.autoAccountID = accountLst[0].AccountID + 1
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
func (self *TAccountMgr) _GetNextAccountID() (ret uint32) {
	ret = self.autoAccountID
	self.mutex.Lock()
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
