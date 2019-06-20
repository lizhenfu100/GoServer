package account

import (
	"common"
	"common/format"
	"common/std/random"
	"common/timer"
	"crypto/md5"
	"dbmgo"
	"fmt"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"gopkg.in/mgo.v2/bson"
	"netConfig/meta"
	"nets/http"
	"shared_svr/svr_center/account/gameInfo"
	"sync/atomic"
	"time"
)

const (
	KDBTable    = "Account"
	kLivelyTime = 1 * 24 * 3600
)

type TAccount struct {
	AccountID  uint32 `bson:"_id"`
	Name       string //账户名
	Password   string //密码 //FIXME:可以StrHash后存成uint32，省不少字节
	CreateTime int64
	LoginTime  int64

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
	return self.Password == fmt.Sprintf("%x", md5.Sum(common.S2B(passwd)))
}
func (self *TAccount) SetPasswd(passwd string) {
	self.Password = fmt.Sprintf("%x", md5.Sum(common.S2B(passwd)))
}

// ------------------------------------------------------------
// 注册、登录
func Rpc_center_account_login(req, ack *common.NetPack) {
	gameName := req.ReadString()
	str := req.ReadString() //账号、邮箱均可登录
	passwd := req.ReadString()

	p := GetAccount(str, passwd)
	errcode := p.Login(passwd)

	if ack.WriteUInt16(errcode); errcode == err.Success {
		ack.WriteUInt32(p.AccountID)
		//附带的游戏数据，可能有的游戏空的
		if v, ok := p.GameInfo[gameName]; ok {
			v.DataToBuf(ack)
		}
	}
}
func (self *TAccount) Login(passwd string) (errcode uint16) {
	if self == nil {
		errcode = err.Account_mismatch_passwd
	} else if !self.CheckPasswd(passwd) {
		errcode = err.Passwd_err
	} else {
		errcode = err.Success
		timeNow := time.Now().Unix()
		atomic.StoreInt64(&self.LoginTime, timeNow)
		dbmgo.UpdateId(KDBTable, self.AccountID, bson.M{"$set": bson.M{"logintime": timeNow}})
		timer.G_TimerMgr.AddTimerSec(func() {
			if time.Now().Unix()-atomic.LoadInt64(&self.LoginTime) >= 15*60 {
				DelCache(self)
			}
		}, 15*60, 0, 0)
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
	} else if GetAccountByBindInfo("email", name) != nil {
		ack.WriteUInt16(err.Account_repeat)
	} else if NewAccountInDB(name, passwd) == nil {
		ack.WriteUInt16(err.Account_repeat)
	} else {
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_center_account_reg_force(req, ack *common.NetPack) {
	uuid := req.ReadString()

	if NewAccountInDB(uuid, "") == nil {
		ack.WriteUInt16(err.Account_repeat)
	} else {
		ack.WriteUInt16(err.Success)
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
	} else if p := GetAccount(name, oldpasswd); p == nil {
		ack.WriteUInt16(err.Account_mismatch_passwd)
	} else {
		p.SetPasswd(newpasswd)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{
			"password": p.Password}})
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_center_create_visitor(req, ack *common.NetPack) {
	name := fmt.Sprintf("%s%d%s",
		random.String(3),
		dbmgo.GetNextIncId("VisitorId"),
		random.String(3))

	if p := NewAccountInDB(name, ""); p == nil {
		name = ""
	}
	ack.WriteString(name)
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

	if p := GetAccountByName(accountName); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		p.WriteLoginAddr(gameName, ack)
	}
}
func (self *TAccount) WriteLoginAddr(gameName string, ack *common.NetPack) {
	if info, ok := self.GameInfo[gameName]; !ok {
		ack.WriteUInt16(err.Not_found)
	} else if p := meta.GetMeta("login", info.LoginSvrId); p == nil {
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
			ids, metas := []int{}, map[int]meta.Meta{}
			for cnt, i := recvBuf.ReadByte(), byte(0); i < cnt; i++ {
				svrId := recvBuf.ReadInt()
				outip := recvBuf.ReadString()
				port := recvBuf.ReadUInt16()
				svrName := recvBuf.ReadString()

				ids = append(ids, svrId) //收集game节点信息，确定具体哪个分流节点
				metas[svrId] = meta.Meta{
					SvrID:   svrId,
					OutIP:   outip,
					TcpPort: port,
					SvrName: svrName,
				}
			}
			if gameInfo.ShuntSvr(ids, &info.GameSvrId, self.AccountID) {
				ack.WriteString(metas[info.GameSvrId].OutIP)
				ack.WriteUInt16(metas[info.GameSvrId].TcpPort)
			}
		})
	}
}
