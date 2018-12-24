package account

import (
	"common"
	"common/format"
	"crypto/md5"
	"dbmgo"
	"fmt"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"netConfig/meta"
	"shared_svr/svr_center/gameInfo"
	"strconv"
	"sync/atomic"
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

	BindInfo map[string]string //email、phone、qq、wechat

	// 有需要可参考player的iModule改写
	GameInfo map[string]gameInfo.TGameInfo
}

func _NewAccount() *TAccount {
	self := new(TAccount)
	self.init()
	return self
}
func (self *TAccount) init() {
	if self.BindInfo == nil {
		self.BindInfo = make(map[string]string, 5)
	}
	if self.GameInfo == nil {
		self.GameInfo = make(map[string]gameInfo.TGameInfo)
	}
}
func (self *TAccount) CheckPasswd(passwd string) bool {
	return self.Password == fmt.Sprintf("%x", md5.Sum([]byte(passwd)))
}
func (self *TAccount) SetPasswd(passwd string) {
	self.Password = fmt.Sprintf("%x", md5.Sum([]byte(passwd)))
}

// ------------------------------------------------------------
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
			if v, ok := account.GameInfo[gameName]; ok {
				v.DataToBuf(ack)
			}
		}
	}
}
func (self *TAccount) Login(passwd string) (errcode uint16) {
	if self == nil {
		errcode = err.Account_none
	} else if !self.CheckPasswd(passwd) {
		errcode = err.Passwd_err
	} else if self.IsForbidden {
		errcode = err.Account_forbidden
	} else {
		errcode = err.Success
		timeNow := time.Now().Unix()
		atomic.StoreInt64(&self.LoginTime, timeNow)
		dbmgo.UpdateId(kDBTable, self.AccountID, bson.M{"$set": bson.M{"logintime": timeNow}})
		time.AfterFunc(15*time.Minute, func() {
			if time.Now().Unix()-atomic.LoadInt64(&self.LoginTime) >= 15*60 {
				DelCache(self)
			}
		})
	}
	return
}
func Rpc_center_account_reg(req, ack *common.NetPack) {
	name := req.ReadString()
	passwd := req.ReadString()

	if !format.CheckAccount(name) {
		ack.WriteUInt16(err.Account_format_err)
	} else if !format.CheckPasswd(passwd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if account := AddNewAccount(name, passwd); account != nil {
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Account_repeat)
	}
}
func Rpc_center_account_check(req, ack *common.NetPack) {
	name := req.ReadString()

	if !format.CheckAccount(name) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if dbmgo.Find(kDBTable, "name", name, &TAccount{}) {
		ack.WriteUInt16(err.Account_repeat)
	} else {
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_center_change_password(req, ack *common.NetPack) {
	name := req.ReadString()
	oldpasswd := req.ReadString()
	newpasswd := req.ReadString()

	if !format.CheckPasswd(newpasswd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if account := GetAccountByName(name); account == nil {
		ack.WriteUInt16(err.Account_none)
	} else if !account.CheckPasswd(oldpasswd) {
		ack.WriteUInt16(err.Passwd_err)
	} else {
		account.SetPasswd(newpasswd)
		dbmgo.UpdateId(kDBTable, account.AccountID, bson.M{"$set": bson.M{
			"password": newpasswd}})
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_center_create_visitor(req, ack *common.NetPack) {
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

// ------------------------------------------------------------
// 记录于账号上面的游戏信息，一套账号系统可关联多个游戏
func Rpc_center_set_game_info(req, ack *common.NetPack) {
	accountId := req.ReadUInt32()
	gameName := req.ReadString()

	if account := GetAccountById(accountId); account != nil {
		var v gameInfo.TGameInfo
		v.BufToData(req)
		account.GameInfo[gameName] = v
		dbmgo.UpdateId(kDBTable, accountId, bson.M{"$set": bson.M{
			fmt.Sprintf("gameinfo.%s", gameName): v}})
	}
}

// 玩家在哪个大区登录的
func Rpc_center_player_login_addr(req, ack *common.NetPack) {
	gameName := req.ReadString()
	accountName := req.ReadString()

	if account := GetAccountByName(accountName); account != nil {
		if v, ok := account.GameInfo[gameName]; ok {
			if p := meta.GetMeta("login", v.LoginSvrId); p == nil {
				ack.WriteUInt16(err.Svr_not_working) //玩家有对应的登录服，但该服未启动
				return
			} else {
				ack.WriteUInt16(err.Success)
				ack.WriteString(p.OutIP)
				ack.WriteUInt16(p.Port())
				return
			}
		}
	}
	ack.WriteUInt16(err.Not_found) //玩家没有对应的登录服，应选取之
}
