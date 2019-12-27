package account

import (
	"common/format"
	"dbmgo"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)

// 账号活跃量很大，预加载内存占用过大

func NewAccountInDB(passwd, bindKey, bindVal string) (uint16, *TAccount) {
	if ok, e := dbmgo.FindEx(KDBTable, bson.M{"$or": []bson.M{
		{"bindinfo.email": bindVal},
		{"bindinfo.name": bindVal},
		{"bindinfo.phone": bindVal},
	}}, &TAccount{}); ok {
		return err.Account_repeat, nil
	} else if e == nil {
		p := _NewAccount()
		p.BindInfo[bindKey] = bindVal
		p.SetPasswd(passwd)
		p.CreateTime = time.Now().Unix()
		p.AccountID = dbmgo.GetNextIncId("AccountId")
		if dbmgo.DB().C(KDBTable).Insert(p) == nil {
			CacheAdd(bindKey+bindVal, p)
			return err.Success, p
		}
	}
	return err.Unknow_error, nil
}
func GetAccountByBindInfo(k, v string) *TAccount { //email、name、phone
	if p := CacheGet(k + v); p != nil {
		return p
	} else {
		p := _NewAccount()
		if ok, _ := dbmgo.Find(KDBTable, "bindinfo."+k, v, p); ok {
			CacheAdd(k+v, p)
			return p
		}
	}
	return nil
}
func GetAccountById(accountId uint32) *TAccount {
	aid := strconv.FormatInt(int64(accountId), 10)
	if p := CacheGet(aid); p != nil {
		return p
	} else {
		p := _NewAccount()
		if ok, _ := dbmgo.Find(KDBTable, "_id", accountId, p); ok {
			CacheAdd(aid, p)
			return p
		}
	}
	return nil
}
func GetAccount(v, passwd string) (uint16, *TAccount) {
	//1、优先当邮箱处理
	if format.CheckBindValue("email", v) {
		p := GetAccountByBindInfo("email", v)
		if p != nil && p.CheckPasswd(passwd) {
			delAccountName(p) //FIXME：删name、bindinfo.name字段
			return err.Success, p
		}
	}
	//2、再当账号名处理
	p := GetAccountByBindInfo("name", v)
	if p != nil && p.CheckPasswd(passwd) {
		moveAccountName(p) //FIXME：删name、bindinfo.name字段，移至bindinfo.email
		return err.Success, p
	}
	return err.Account_mismatch_passwd, nil
}

// ------------------------------------------------------------
// FIXME：修补上古遗祸~囧 2019.12.25
func delAccountName(p *TAccount) {
	if format.CheckBindValue("email", p.BindInfo["name"]) {
		gamelog.Info("delAccountName: %s %s", p.Name, p.BindInfo["name"])
		p.Name = ""
		p.BindInfo["name"] = ""
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$unset": bson.M{"name": 1, "bindinfo.name": 1}})
	}
}
func moveAccountName(p *TAccount) {
	if format.CheckBindValue("email", p.BindInfo["name"]) {
		if GetAccountByBindInfo("email", p.BindInfo["name"]) == nil {
			gamelog.Info("moveAccountName: %s %s %s", p.Name, p.BindInfo["name"], p.BindInfo["email"])
			p.Name = ""
			p.BindInfo["email"] = p.BindInfo["name"]
			p.BindInfo["name"] = ""
			dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{
				"$unset": bson.M{"name": 1, "bindinfo.name": 1},
				"$set":   bson.M{"bindinfo.email": p.BindInfo["email"]},
			})
		}
	}
}
