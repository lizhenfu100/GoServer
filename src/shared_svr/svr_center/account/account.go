package account

import (
	"common"
	"common/format"
	"dbmgo"
	"fmt"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"shared_svr/svr_center/gameInfo/HappyDiner"
	"shared_svr/svr_center/gameInfo/SoulKnight"
	"strconv"
	"time"
)

const kDBTable = "Account"

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
	HappyDiner HappyDiner.TGameInfo
}

func _NewAccount() *TAccount {
	self := new(TAccount)
	self.init()
	return self
}
func (self *TAccount) init() {
	self.gameList = map[string]IGameInfo{
		"SoulKnight": &self.SoulKnight,
		"HappyDiner": &self.HappyDiner,
	}
}
func (self *TAccount) Login(passwd string) (errcode uint16) {
	if self == nil {
		errcode = err.Account_none
	} else if passwd != self.Password {
		errcode = err.Passwd_err
	} else if self.IsForbidden {
		errcode = err.Account_forbidden
	} else {
		errcode = err.Success

		self.LoginTime = time.Now().Unix()
		dbmgo.UpdateToDB(kDBTable, bson.M{"_id": self.AccountID}, bson.M{"$set": bson.M{"logintime": self.LoginTime}})
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
		ack.WriteUInt16(err.Account_none)
	} else {
		errcode := account.Login(passwd)
		ack.WriteUInt16(errcode)
		if errcode == err.Success {
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
		ack.WriteUInt16(err.Account_format_err)
	} else if !format.CheckPasswd(password) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if account := AddNewAccount(name, password); account != nil {
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Account_repeat)
	}
}
func Rpc_center_check_account(req, ack *common.NetPack) {
	name := req.ReadString()

	if !format.CheckAccount(name) {
		ack.WriteUInt16(err.Passwd_format_err)
		return
	}
	var account TAccount
	if dbmgo.Find(kDBTable, "name", name, &account) {
		ack.WriteUInt16(err.Account_repeat)
	} else {
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_center_change_password(req, ack *common.NetPack) {
	name := req.ReadString()
	oldpassword := req.ReadString()
	newpassword := req.ReadString()

	if !format.CheckPasswd(newpassword) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if ok := ResetPassword(name, oldpassword, newpassword); ok {
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Passwd_err)
	}
}
func Rpc_center_visitor_account(req, ack *common.NetPack) {
	id := dbmgo.GetNextIncId("VisitorId")
	name := fmt.Sprintf("ChillyRoomGuest_%d", id)
	passwd := strconv.Itoa(int(common.StringHash(name)))

	if account := AddNewAccount(name, passwd); account == nil {
		gamelog.Error("visitor_account fail: %s:%s", name, passwd)
		name = ""
		passwd = ""
	}
	ack.WriteString(name)
	ack.WriteString(passwd)
}
