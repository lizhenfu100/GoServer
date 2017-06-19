package account

import (
	"common"
	"dbmgo"
	"net/http"
	"netConfig"
	"svr_center/api"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type TAccount struct {
	AccountID   uint32 `bson:"_id"` //账号ID
	Name        string //账户名
	Password    string //密码
	CreateTime  int64
	LoginTime   int64
	LoginCount  uint32
	LoginSvrID  uint32
	IsForbidden bool //是否禁用
}

//处理用户账户注册请求
func Rpc_Reg_Account(req, ack *common.NetPack, ptr interface{}) {
	name := req.ReadString()
	password := req.ReadString()

	if account := G_AccountMgr.AddNewAccount(name, password); account != nil {
		ack.WriteInt8(1)
	} else {
		ack.WriteInt8(-1)
	}
}
func Rpc_Change_Password(req, ack *common.NetPack, ptr interface{}) {
	name := req.ReadString()
	oldpassword := req.ReadString()
	newpassword := req.ReadString()

	if ok := G_AccountMgr.ResetPassword(name, oldpassword, newpassword); ok {
		ack.WriteInt8(1)
	} else {
		ack.WriteInt8(-1)
	}
}
func Rpc_GetGameSvr_Lst(req, ack *common.NetPack, ptr interface{}) {
	cfgLst := api.GetRegGamesvrCfgLst()
	ack.WriteByte(byte(len(cfgLst)))
	for _, v := range cfgLst {
		ack.WriteUInt32(uint32(v.SvrID))
		ack.WriteString(v.SvrName)
		ack.WriteString(v.OutIP)
		if v.HttpPort > 0 {
			ack.WriteUInt16(uint16(v.HttpPort))
		} else {
			ack.WriteUInt16(uint16(v.TcpPort))
		}
	}
}
func Rpc_GetGameSvr_LastLogin(req, ack *common.NetPack, ptr interface{}) {
	name := req.ReadString()

	account := G_AccountMgr.GetAccountByName(name)
	if account == nil {
		ack.WriteInt8(-1) //not_exist
	} else {
		ack.WriteInt8(1)
		ack.WriteUInt32(account.LoginSvrID)

		svrId := int(account.LoginSvrID)
		if cfg := netConfig.GetNetCfg("game", &svrId); cfg != nil {
			ack.WriteString(cfg.SvrName)
		}
	}
}
func Rpc_Login_GameSvr(req, ack *common.NetPack, ptr interface{}) {
	name := req.ReadString()
	password := req.ReadString()
	svrId := req.ReadInt()

	account := G_AccountMgr.GetAccountByName(name)
	if account == nil {
		ack.WriteInt8(-1) //not_exist
	} else if account.IsForbidden {
		ack.WriteInt8(-2) //forbidded_account
	} else if password != account.Password {
		ack.WriteInt8(-3) //invalid_password
	} else if netConfig.GetNetCfg("game", &svrId) == nil {
		ack.WriteInt8(-4) //forbidded_account
	} else {
		ack.WriteInt8(1)
		ack.WriteUInt32(account.AccountID)
		netConfig.WriteAddr(ack, "game", &svrId)

		//生成一个临时token，发给gamesvr、client，用以登录验证
		token := G_AccountMgr.CreateLoginToken()
		ack.WriteUInt32(token)
		buf := common.NewByteBufferCap(8)
		buf.WriteUInt32(account.AccountID)
		buf.WriteUInt32(token)
		api.SendToGame(svrId, "login_token", buf.DataPtr)
	}
}
func Handle_Login_Game_Success(w http.ResponseWriter, r *http.Request) {
	//! 接收信息
	req := common.NewByteBufferLen(int(r.ContentLength))
	r.Body.Read(req.DataPtr)

	accountId := req.ReadUInt32()
	svrId := req.ReadUInt32()

	if account := G_AccountMgr.GetAccountById(accountId); account != nil {
		account.LoginCount++
		account.LoginSvrID = svrId
		account.LoginTime = time.Now().Unix()
		dbmgo.UpdateToDB("Account", bson.M{"_id": accountId}, bson.M{"$set": bson.M{
			"loginsvrid": svrId,
			"logincount": account.LoginCount,
			"logintime":  account.LoginTime}})
	}
}
