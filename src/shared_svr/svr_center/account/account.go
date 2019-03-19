package account

import (
	"common"
	"common/format"
	"common/std/hash"
	"crypto/md5"
	"dbmgo"
	"fmt"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"gopkg.in/mgo.v2/bson"
	"http"
	"netConfig/meta"
	"shared_svr/svr_center/gameInfo"
	"strconv"
	"sync/atomic"
	"time"
)

const KDBTable = "Account"

type TAccount struct {
	AccountID   uint32 `bson:"_id"`
	Name        string //账户名
	Password    string //密码 //FIXME:可以StrHash后存成uint32，省不少字节
	CreateTime  int64
	LoginTime   int64
	ForbidTime  int64
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
	return self.Password == fmt.Sprintf("%x", md5.Sum(common.ToBytes(passwd)))
}
func (self *TAccount) SetPasswd(passwd string) {
	self.Password = fmt.Sprintf("%x", md5.Sum(common.ToBytes(passwd)))
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
		dbmgo.UpdateId(KDBTable, self.AccountID, bson.M{"$set": bson.M{"logintime": timeNow}})
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
		ack.WriteUInt16(err.Account_format_err)
	} else if ok, _ := dbmgo.Find(KDBTable, "name", name, &TAccount{}); ok {
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
		dbmgo.UpdateId(KDBTable, account.AccountID, bson.M{"$set": bson.M{
			"password": account.Password}})
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_center_create_visitor(req, ack *common.NetPack) {
	id := dbmgo.GetNextIncId("VisitorId")
	name := fmt.Sprintf("ChillyRoomGuest_%d", id)
	passwd := strconv.Itoa(int(hash.StrHash(name)))

	if account := AddNewAccount(name, passwd); account == nil {
		gamelog.Error("visitor_account fail: %s:%s", name, passwd)
		name = ""
		passwd = ""
	}
	ack.WriteString(name)
	ack.WriteString(passwd)
}
func Rpc_center_account_forbid(req, ack *common.NetPack) {
	id := req.ReadUInt32()

	if p := GetAccountById(id); p != nil && !G_WhiteList.Have(p.Name) {
		p.IsForbidden = true
		p.ForbidTime = time.Now().Unix()
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{
			"forbidtime":  p.ForbidTime,
			"isforbidden": true}})
	}
}

// ------------------------------------------------------------
// 记录于账号上面的游戏信息，一套账号系统可关联多个游戏
func Rpc_center_set_game_info(req, ack *common.NetPack) {
	accountId := req.ReadUInt32()
	gameName, info := req.ReadString(), gameInfo.TGameInfo{}
	info.BufToData(req)

	if account := GetAccountById(accountId); account != nil && gameName != "" {
		account.GameInfo[gameName] = info
		if dbmgo.UpdateIdSync(KDBTable, accountId, bson.M{"$set": bson.M{
			fmt.Sprintf("gameinfo.%s", gameName): info}}) {
			ack.WriteUInt16(err.Success)
			return
		}
	}
	ack.WriteUInt16(err.GameInfo_set_fail)
	gamelog.Error("set_game_info: %d %s %v", accountId, gameName, info)
}

// 玩家在哪个大区登录的
func Rpc_center_player_login_addr(req, ack *common.NetPack) {
	gameName := req.ReadString()
	accountName := req.ReadString()

	if account := GetAccountByName(accountName); account != nil {
		if info, ok := account.GameInfo[gameName]; ok {
			if p := meta.GetMeta("login", info.LoginSvrId); p == nil {
				ack.WriteUInt16(err.Svr_not_working) //玩家有对应的登录服，但该服未启动
			} else {
				ack.WriteUInt16(err.Success)
				ack.WriteString(p.OutIP)
				ack.WriteUInt16(p.Port())

				// 向login查询game的地址，一并回给client
				http.CallRpc(http.Addr(p.OutIP, p.Port()), enum.Rpc_meta_list, func(buf *common.NetPack) {
					buf.WriteString("game")
					buf.WriteString(meta.G_Local.Version)
				}, func(recvBuf *common.NetPack) {
					for cnt, i := recvBuf.ReadByte(), byte(0); i < cnt; i++ {
						svrId := recvBuf.ReadInt()
						outip := recvBuf.ReadString()
						port := recvBuf.ReadUInt16()
						recvBuf.ReadString() //svrName
						gamelog.Debug("GameAddr -- %s:%d(%d)", outip, port, svrId)
						if svrId == info.GameSvrId {
							ack.WriteString(outip)
							ack.WriteUInt16(port)
							break
						}
					}
				})
			}
			return
		}
	}
	ack.WriteUInt16(err.Not_found) //玩家没有对应的登录服，应选取之
}
