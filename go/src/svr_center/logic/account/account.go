package account

import (
	"common"
	"dbmgo"
	"gamelog"
	"net/http"
	"svr_center/api"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//处理用户账户注册请求
func Rpc_Reg_Account(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	req := common.NewNetPackLen(int(r.ContentLength))
	r.Body.Read(req.DataPtr)

	name := req.ReadString()
	password := req.ReadString()

	//! 创建回复
	ack := common.NewNetPackCap(64)
	defer w.Write(ack.DataPtr)

	if ptr := G_AccountMgr.AddNewAccount(name, password); ptr != nil {
		ack.WriteInt8(1)
	} else {
		ack.WriteInt8(-1)
	}
}
func Rpc_GetGameSvrLst(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	req := common.NewNetPackLen(int(r.ContentLength))
	r.Body.Read(req.DataPtr)

	name := req.ReadString()
	password := req.ReadString()

	//! 创建回复
	ack := common.NewNetPackCap(64)
	defer w.Write(ack.DataPtr)

	account := G_AccountMgr.GetAccountByName(name)
	if account == nil {
		ack.WriteInt8(-1) //not_exist
	} else if account.Forbidden {
		ack.WriteInt8(-2) //forbidded_account
	} else if password == account.Password {
		ack.WriteInt8(0)
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
func Rpc_Login_Success(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	req := common.NewNetPackLen(int(r.ContentLength))
	r.Body.Read(req.DataPtr)

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
