package account

import (
	"common"
	"dbmgo"
	"svr_center/api"
	"time"

	"gopkg.in/mgo.v2/bson"
)

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

//处理用户账户注册请求
func Rpc_Reg_Account(req, ack *common.ByteBuffer) {
	name := req.ReadString()
	password := req.ReadString()

	if ptr := G_AccountMgr.AddNewAccount(name, password); ptr != nil {
		ack.WriteInt8(1)
	} else {
		ack.WriteInt8(-1)
	}
}
func Rpc_GetGameSvrLst(req, ack *common.ByteBuffer) {
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
		//TODO:zhoumf:生成一个临时token，发给gamesvr、client，用以登录验证
		// token := CreateLoginToken()
		// ack.WriteString(token)
		// api.RelayToGamesvr(1, strKey, token)
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
func Rpc_Login_Success(req, ack *common.ByteBuffer) {
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
