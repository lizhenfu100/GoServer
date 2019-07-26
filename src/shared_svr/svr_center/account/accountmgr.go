package account

import (
	"dbmgo"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	g_aid_cache  sync.Map //map[uint32]*TAccount
	g_bind_cache sync.Map //map[string]*TAccount
)

func InitDB() {
	var list []TAccount //只载入近期登录过的
	//dbmgo.FindAll(KDBTable, bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - kActiveTime}}, &list)
	//for i := 0; i < len(list); i++ {
	//	list[i].init()
	//	AddCache(&list[i])
	//}
	println("load active account form db: ", len(list))
}
func NewAccountInDB(passwd, bindKey, bindVal string) (uint16, *TAccount) {
	if ok, _ := dbmgo.FindEx(KDBTable, bson.M{"$or": []bson.M{
		{"bindinfo.email": bindVal},
		{"bindinfo.name": bindVal},
		{"bindinfo.phone": bindVal},
	}}, &TAccount{}); ok {
		return err.Account_repeat, nil
	}
	account := _NewAccount()
	account.BindInfo[bindKey] = bindVal
	account.SetPasswd(passwd)
	account.CreateTime = time.Now().Unix()
	account.AccountID = dbmgo.GetNextIncId("AccountId")
	if dbmgo.InsertSync(KDBTable, account) {
		AddCache(account)
		return err.Success, account
	}
	return err.Unknow_error, nil
}
func GetAccountByBindInfo(k, v string) *TAccount {
	if p, ok := g_bind_cache.Load("bindinfo." + k + v); ok {
		return p.(*TAccount)
	} else {
		account := _NewAccount()
		if ok, _ := dbmgo.Find(KDBTable, "bindinfo."+k, v, account); ok {
			AddCache(account)
			return account
		}
	}
	return nil
}
func GetAccountById(accountId uint32) *TAccount {
	if v, ok := g_aid_cache.Load(accountId); ok {
		return v.(*TAccount)
	} else {
		account := _NewAccount()
		if ok, _ := dbmgo.Find(KDBTable, "_id", accountId, account); ok {
			AddCache(account)
			return account
		}
	}
	return nil
}
func GetAccount(v, passwd string) (uint16, *TAccount) {
	//1、优先当邮箱处理
	p1 := GetAccountByBindInfo("email", v)
	if p1 != nil && p1.CheckPasswd(passwd) {
		return err.Success, p1
	}
	//2、再当账号名处理
	p2 := GetAccountByBindInfo("name", v)
	if p2 != nil && p2.CheckPasswd(passwd) {
		return err.Success, p2
	}
	if p1 == nil && p2 == nil {
		return err.Not_found, nil
	} else {
		return err.Account_mismatch_passwd, nil
	}
}

// ------------------------------------------------------------
//! 辅助函数
func AddCache(p *TAccount) {
	if p.Name != "" && p.BindInfo["name"] == "" { //TODO:待删除
		p.BindInfo["name"] = p.Name
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"bindinfo.name": p.Name}})
	}
	g_aid_cache.Store(p.AccountID, p)
	for k, v := range p.BindInfo {
		g_bind_cache.Store("bindinfo."+k+v, p)
	}
}
func DelCache(p *TAccount) {
	g_aid_cache.Delete(p.AccountID)
	for k, v := range p.BindInfo {
		g_bind_cache.Delete("bindinfo." + k + v)
	}
}
