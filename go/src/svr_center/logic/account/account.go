package account

import (
	"common"
	"common/format"
	"dbmgo"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type TAccount struct {
	AccountID   uint32 `bson:"_id"` //账号ID
	Name        string //账户名
	Password    string //密码
	CreateTime  int64
	LoginTime   int64
	IsForbidden bool //是否禁用

	BindInfo map[string]string //邮箱、手机、微信号

	// GameList []int //登录过的游戏svrId列表，svrId格式：前四位-游戏种类ID、后四位-区服ID
}

func (self *TAccount) Login(passwd string) (errcode int8) {
	if self == nil {
		errcode = -1 //not_exist
	} else if passwd != self.Password {
		errcode = -2 //invalid_password
	} else if self.IsForbidden {
		errcode = -3 //forbidded_account
	} else {
		errcode = 1

		self.LoginTime = time.Now().Unix()
		dbmgo.UpdateToDB("Account", bson.M{"_id": self.AccountID}, bson.M{"$set": bson.M{"logintime": self.LoginTime}})

		time.AfterFunc(15*time.Minute, func() {
			G_AccountMgr.DelToCache(self)
		})
	}
	return
}

// -------------------------------------
// 注册、登录
func Rpc_center_reg_account(req, ack *common.NetPack) {
	name := req.ReadString()
	password := req.ReadString()

	if !format.CheckAccount(name) {
		ack.WriteInt8(-1)
	} else if !format.CheckPasswd(password) {
		ack.WriteInt8(-2)
	} else if account := G_AccountMgr.AddNewAccount(name, password); account != nil {
		ack.WriteInt8(1)
	} else {
		ack.WriteInt8(-3)
	}
}
func Rpc_center_check_account(req, ack *common.NetPack) {
	name := req.ReadString()

	if !format.CheckAccount(name) {
		ack.WriteInt8(-1)
		return
	}
	var account TAccount
	if ok := dbmgo.Find("Account", "name", name, &account); ok {
		ack.WriteInt8(-2)
	} else {
		ack.WriteInt8(1)
	}
}
func Rpc_center_change_password(req, ack *common.NetPack) {
	name := req.ReadString()
	oldpassword := req.ReadString()
	newpassword := req.ReadString()

	if !format.CheckPasswd(newpassword) {
		ack.WriteInt8(-1)
	} else if ok := G_AccountMgr.ResetPassword(name, oldpassword, newpassword); ok {
		ack.WriteInt8(1)
	} else {
		ack.WriteInt8(-2)
	}
}
func Rpc_center_account_login(req, ack *common.NetPack) {
	name := req.ReadString()
	passwd := req.ReadString()

	account := G_AccountMgr.GetAccountByName(name)
	errcode := account.Login(passwd)
	ack.WriteInt8(errcode)
	if errcode > 0 {
		ack.WriteUInt32(account.AccountID)
	}
}
