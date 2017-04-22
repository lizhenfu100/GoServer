package account

import (
	"common"
	"dbmgo"
	"svr_center/api"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//处理用户账户注册请求
func Rpc_Reg_Account(req, ack *common.NetPack) {
	name := req.ReadString()
	password := req.ReadString()

	if ptr := G_AccountMgr.AddNewAccount(name, password); ptr != nil {
		ack.WriteInt8(1)
	} else {
		ack.WriteInt8(-1)
	}
}
func Rpc_GetGameSvrLst(req, ack *common.NetPack) {
	name := req.ReadString()
	password := req.ReadString()

	account := G_AccountMgr.GetAccountByName(name)
	if account == nil {
		ack.WriteInt8(-1) //not_exist
	} else if account.Forbidden {
		ack.WriteInt8(-2) //forbidded_account
	} else if password == account.Password {
		ack.WriteInt8(1)
		ack.WriteUint32(account.AccountID)
		ack.WriteUint32(account.LoginSvrID)
		//游戏服列表
		cfgLst := api.GetRegGamesvrCfgLst()
		ack.WriteByte(byte(len(cfgLst)))
		for _, v := range cfgLst {
			ack.WriteString(v.Module)
			ack.WriteUint32(uint32(v.SvrID))
			ack.WriteString(v.OutIP)
			ack.WriteUint16(uint16(v.HttpPort))
		}
	} else {
		ack.WriteInt8(-3) //invalid_password
	}
}
func Rpc_Login_Success(req, ack *common.NetPack) {
	accountId := req.ReadUint32()
	svrId := req.ReadUint32()

	if account, ok := G_AccountMgr.IdToPtr[accountId]; ok {
		account.LoginCount++
		account.LoginSvrID = svrId
		account.LoginTime = time.Now().Unix()
		dbmgo.UpdateToDB("Account", bson.M{"_id": accountId}, bson.M{"$set": bson.M{
			"loginsvrid": svrId,
			"logincount": account.LoginCount,
			"logintime":  account.LoginTime}})
	}
}
