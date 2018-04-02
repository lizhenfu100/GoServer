package account

import (
	"common"
	"common/format"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"shared_svr/svr_center/gameInfo/SoulKnight"
	"time"
)

type TAccount struct {
	AccountID   uint32 `bson:"_id"`
	Name        string //账户名
	Password    string //密码
	CreateTime  int64
	LoginTime   int64
	IsForbidden bool //是否禁用

	BindInfo map[string]string //邮箱、手机、微信号

	/* --- db game info --- */
	// 有需要可参考player的iDBModule改写
	gameList   map[string]IGameInfo
	SoulKnight SoulKnight.TGameInfo
}

func _NewAccount() *TAccount {
	self := new(TAccount)
	self.init()
	return self
}
func (self *TAccount) init() {
	self.gameList = map[string]IGameInfo{
		"SoulKnight": &self.SoulKnight,
	}
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
			if time.Now().Unix()-self.LoginTime >= 15*60 {
				DelCache(self)
			}
		})
	}
	return
}
func (self *TAccount) GetGameInfo(gameName string) IGameInfo {
	if p, ok := self.gameList[gameName]; ok {
		return p
	}
	return nil
}

// -------------------------------------
// 注册、登录
func Rpc_center_account_login(req, ack *common.NetPack) {
	gameName := req.ReadString()
	name := req.ReadString()
	passwd := req.ReadString()

	if account := GetAccountByName(name); account == nil {
		ack.WriteInt8(-1)
	} else {
		errcode := account.Login(passwd)
		ack.WriteInt8(errcode)
		if errcode > 0 {
			ack.WriteUInt32(account.AccountID)
			//附带的游戏数据，可能有的游戏空的
			if pInfo := account.GetGameInfo(gameName); pInfo != nil {
				pInfo.DataToBuf(ack)
			}
		}
	}
}
func Rpc_center_reg_account(req, ack *common.NetPack) {
	name := req.ReadString()
	password := req.ReadString()

	if !format.CheckAccount(name) {
		ack.WriteInt8(-1)
	} else if !format.CheckPasswd(password) {
		ack.WriteInt8(-2)
	} else if account := AddNewAccount(name, password); account != nil {
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
	if dbmgo.Find("Account", "name", name, &account) {
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
	} else if ok := ResetPassword(name, oldpassword, newpassword); ok {
		ack.WriteInt8(1)
	} else {
		ack.WriteInt8(-2)
	}
}
