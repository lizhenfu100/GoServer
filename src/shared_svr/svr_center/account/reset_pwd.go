package account

import (
	"common"
	"common/format"
	"dbmgo"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
)

func Rpc_center_ask_reset_password(req, ack *common.NetPack) {
	name := req.ReadString()

	if account := GetAccountByName(name); account == nil {
		ack.WriteUInt16(err.Account_none)
	} else if addr, ok := account.BindInfo["email"]; !ok {
		ack.WriteUInt16(err.Account_without_bind_info)
	} else {
		ack.WriteUInt16(err.Success)
		ack.WriteUInt32(account.AccountID)
		ack.WriteString(addr)
	}
}
func Rpc_center_reset_password(req, ack *common.NetPack) {
	aid := req.ReadUInt32()
	passwd := req.ReadString()

	if !format.CheckPasswd(passwd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if account := GetAccountById(aid); account == nil {
		ack.WriteUInt16(err.Account_none)
	} else {
		account.SetPasswd(passwd)
		dbmgo.UpdateIdToDB(kDBTable, account.AccountID, bson.M{"$set": bson.M{
			"password": passwd}})
		ack.WriteUInt16(err.Success)
	}
}
