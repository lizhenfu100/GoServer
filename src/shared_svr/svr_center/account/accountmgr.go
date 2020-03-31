package account

import (
	"common"
	"common/format"
	"dbmgo"
	"encoding/json"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)

func NewAccountInDB(passwd, bindTyp, bindVal string) (uint16, *TAccount) {
	if ok, e := dbmgo.FindEx(KDBTable, bson.M{"$or": []bson.M{
		{"bindinfo.email": bindVal},
		{"bindinfo.name": bindVal},
		{"bindinfo.phone": bindVal},
	}}, &TAccount{}); ok {
		return err.Account_repeat, nil
	} else if e == nil {
		p := _NewAccount()
		p.BindInfo[bindTyp] = bindVal
		p.SetPasswd(passwd)
		p.CreateTime = time.Now().Unix()
		p.AccountID = dbmgo.GetNextIncId("AccountId")
		if dbmgo.DB().C(KDBTable).Insert(p) == nil {
			CacheAdd(bindTyp+bindVal, p)
			return err.Success, p
		}
	}
	return err.Unknow_error, nil
}
func GetAccountByBindInfo(typ, val string) (uint16, *TAccount) { //email、name、phone
	if p := CacheGet(typ + val); p != nil {
		return err.Success, p
	} else {
		p = _NewAccount()
		if ok, e := dbmgo.Find(KDBTable, "bindinfo."+typ, val, p); ok {
			CacheAdd(typ+val, p)
			return err.Success, p
		} else if e == nil {
			return err.Not_found, nil
		}
		return err.Unknow_error, nil
	}
}
func GetAccountById(id uint32) (uint16, *TAccount) {
	aid := strconv.FormatInt(int64(id), 10)
	if p := CacheGet(aid); p != nil {
		return err.Success, p
	} else {
		p = _NewAccount()
		if ok, e := dbmgo.Find(KDBTable, "_id", id, p); ok {
			CacheAdd(aid, p)
			return err.Success, p
		} else if e == nil {
			return err.Not_found, nil
		}
		return err.Unknow_error, nil
	}
}
func GetAccount(v, passwd string) (uint16, *TAccount) {
	var p1, p2 *TAccount
	e := err.Unknow_error
	//1、优先当邮箱处理
	if format.CheckBindValue("email", v) {
		if e, p1 = GetAccountByBindInfo("email", v); e == err.Unknow_error {
			return e, nil
		} else if p1 != nil && p1.CheckPasswd(passwd) {
			delAccountName(p1) //FIXME：删name、bindinfo.name字段
			return err.Success, p1
		}
	}
	//2、再当账号名处理
	if e, p2 = GetAccountByBindInfo("name", v); e == err.Unknow_error {
		return e, nil
	} else if p2 != nil && p2.CheckPasswd(passwd) {
		moveAccountName(p2) //FIXME：删name、bindinfo.name字段，移至bindinfo.email
		return err.Success, p2
	}
	//if p1 == nil && p2 == nil && format.CheckBindValue("phone", v) {
	//	if e, p1 = GetAccountByBindInfo("phone", v); e == err.Unknow_error {
	//		return e, nil
	//	} else if p1 != nil && p1.CheckPasswd(passwd) {
	//		return err.Success, p1
	//	}
	//}
	if p1 == nil && p2 == nil {
		return err.Not_found, nil
	} else {
		return err.Account_mismatch_passwd, nil
	}
}

// ------------------------------------------------------------
// FIXME：修补上古遗祸~囧 2019.12.25
func delAccountName(p *TAccount) {
	if name := p.BindInfo["name"]; format.CheckBindValue("email", name) {
		gamelog.Info("delAccountName: %s %s", p.Name, name)
		p.Name = ""
		delete(p.BindInfo, "name")
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$unset": bson.M{"name": 1, "bindinfo.name": 1}})
	}
}
func moveAccountName(p *TAccount) {
	if name := p.BindInfo["name"]; format.CheckBindValue("email", name) {
		if _, ptr := GetAccountByBindInfo("email", name); ptr != nil && ptr.AccountID != p.AccountID {
			if !dbmgo.RemoveOneSync(KDBTable, bson.M{"_id": ptr.AccountID}) {
				return
			}
			b, _ := json.Marshal(ptr)
			dbmgo.Log("delaccount", name, common.B2S(b))
		}
		gamelog.Info("moveAccountName: %s %s", p.Name, name)
		p.Name = ""
		delete(p.BindInfo, "name")
		p.BindInfo["email"] = name
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{
			"$unset": bson.M{"name": 1, "bindinfo.name": 1},
			"$set":   bson.M{"bindinfo.email": name},
		})
	}
}
