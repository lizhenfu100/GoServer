package account

import (
	"common"
	"common/format"
	"dbmgo"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

// -------------------------------------
// 绑定信息到账号
func Rpc_center_bind_info(req, ack *common.NetPack) {
	name := req.ReadString()
	passwd := req.ReadString()
	key := req.ReadString()
	val := req.ReadString()
	force := req.ReadInt8()

	errcode := BindInfoToAccount(name, passwd, key, val, force)
	ack.WriteInt8(errcode)
}
func Rpc_center_get_account_by_bind_info(req, ack *common.NetPack) {
	val := req.ReadString()
	key := req.ReadString()
	passwd := req.ReadString()

	account := GetAccountByBindInfo(key, val)
	if account != nil {
		ack.WriteInt8(-1)
	} else if passwd != account.Password {
		ack.WriteInt8(-2)
	} else {
		ack.WriteInt8(1)
		ack.WriteString(account.Name)
	}
}

// -------------------------------------
// 辅助函数
func BindInfoToAccount(name, passwd, k, v string, force int8) (errcode int8) {
	dbkey := fmt.Sprintf("bindinfo.%s", k)
	if account := GetAccountByName(name); account == nil {
		errcode = -1 //not_exist
	} else if passwd != account.Password {
		errcode = -2 //invalid_password
	} else if account.IsForbidden {
		errcode = -3 //forbidded_account
	} else if !format.CheckValue(k, v) {
		errcode = -4 //invalid_value
	} else if dbmgo.Find("Account", dbkey, v, new(TAccount)) {
		errcode = -5 //this_value_already_bind_to_account
	} else {
		if account.BindInfo == nil {
			account.BindInfo = make(map[string]string, 5)
		}
		if _, ok := account.BindInfo[k]; ok && force == 0 {
			errcode = -6 //can_force_cover
		} else {
			errcode = 1
			account.BindInfo[k] = v
			dbmgo.UpdateToDB("Account", bson.M{"_id": account.AccountID}, bson.M{"$set": bson.M{dbkey: v}})
		}
	}
	return
}
func GetAccountByBindInfo(k, v string) *TAccount {
	//FIXME: 数据多了，这样没加索引的找太慢了。可将结果缓存，仅第一次找慢些 "bindinfo.k.v" = ptr
	dbkey := fmt.Sprintf("bindinfo.%s", k)
	account := new(TAccount)
	if dbmgo.Find("Account", dbkey, v, account) {
		return account
	}
	return nil
}
