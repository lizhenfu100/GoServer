package account

import (
	"common"
	"common/format"
	"dbmgo"
	"fmt"

	"generate_out/err"
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
	ack.WriteUInt16(errcode)
}
func Rpc_center_get_account_by_bind_info(req, ack *common.NetPack) { //拿到账号名，client本地保存
	val := req.ReadString()
	key := req.ReadString()
	passwd := req.ReadString()

	account := GetAccountByBindInfo(key, val)
	if account != nil {
		ack.WriteUInt16(err.Account_none)
	} else if !account.CheckPasswd(passwd) {
		ack.WriteUInt16(err.Passwd_err)
	} else {
		ack.WriteUInt16(err.Success)
		ack.WriteString(account.Name)
	}
}

// -------------------------------------
// 辅助函数
func BindInfoToAccount(name, passwd, k, v string, force int8) (errcode uint16) {
	if account := GetAccountByName(name); account == nil {
		errcode = err.Account_none
	} else if !account.CheckPasswd(passwd) {
		errcode = err.Passwd_err
	} else if account.IsForbidden {
		errcode = err.Account_forbidden
	} else if !format.CheckValue(k, v) {
		errcode = err.BindInfo_format_err
	} else if GetAccountByBindInfo(k, v) != nil {
		errcode = err.BindInfo_had_been_bound //this_value_already_bind_to_account
	} else {
		if _, ok := account.BindInfo[k]; ok && force == 0 {
			errcode = err.Account_had_been_bound //can_force_cover
		} else {
			errcode = err.Success
			account.BindInfo[k] = v
			dbkey := fmt.Sprintf("bindinfo.%s", k)
			dbmgo.UpdateId(kDBTable, account.AccountID, bson.M{"$set": bson.M{dbkey: v}})
		}
	}
	return
}
func GetAccountByBindInfo(k, v string) *TAccount {
	//FIXME: 数据多了，这样没加索引的找太慢了
	//客户端通过绑定信息查到账号后，将账号名保存至本地，之后用账户名登录
	dbkey := fmt.Sprintf("bindinfo.%s", k)
	account := new(TAccount)
	if dbmgo.Find(kDBTable, dbkey, v, account) {
		return account
	}
	return nil
}
