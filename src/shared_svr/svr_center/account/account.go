/***********************************************************************
* @ 账号管理

* @ 目前，所有center同质化的，连一个db

* @ 若注册数量到亿级，必须分库，“哈希id”或“id分段”至多个db TODO:optimize
	· 哈希：每次扩容须迁移部分账号
	· 分段：新玩家集中，单点压力大 …… 倾向于JumpHash方式
	· 玩家分别用 “账号名、邮箱、手机号” 登录，如何快速定位节点？
		· <字符串,ID>映射，哈希字符串，同样存到多个db
		· 须保证注册、改名、改邮箱...事务性修改映射表~囧~研究下saga模式
		· 先查映射表，再查账号数据

* @ 要服务全球玩家，访问账号必须加速，各登录节点得有账号redis
	· 读，先读缓存，没有再读center，成功后写缓存（缓存会过期）
	· 写，先更db，再删缓存

* @ author zhoumf
* @ date 2019-8-23
***********************************************************************/
package account

import (
	"common"
	"common/format"
	"common/std/random"
	"common/std/sign"
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
	KDBTable = "Account"
)

type TAccount struct {
	AccountID  uint32 `bson:"_id"`
	Name       string //TODO:待删除
	Password   string //FIXME:可以StrHash后存成uint32，省不少字节
	CreateTime int64
	LoginTime  int64

	//TODO：手机号客户端就做加密，整个流转中途都是加密的
	BindInfo     map[string]string //email、phone、name
	IsValidEmail uint8
	IsValidPhone uint8

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
		self.BindInfo = make(map[string]string, 3)
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
	pwd := req.ReadString()
	sign.Decode(&str, &pwd)
	gamelog.Debug("Login: %s %s", str, pwd)

	errcode, p := GetAccount(str, pwd)
	if ack.WriteUInt16(errcode); errcode == err.Success {
		timeNow := time.Now().Unix()
		atomic.StoreInt64(&p.LoginTime, timeNow)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"logintime": timeNow}})
		//1、先回复，Client账号信息
		v := gameInfo.TAccountClient{
			p.AccountID,
			p.IsValidEmail,
			p.IsValidPhone,
		}
		v.DataToBuf(ack)
		//2、再回复，附带的游戏数据，可能有的游戏空的
		if v, ok := p.GameInfo[gameName]; ok {
			ack.WriteInt(v.LoginSvrId)
			ack.WriteInt(v.GameSvrId)
		}
	}
}
func Rpc_center_account_reg(req, ack *common.NetPack) {
	str := req.ReadString()
	pwd := req.ReadString()
	typ := req.ReadString() //email、name、phone
	sign.Decode(&str, &pwd)
	gamelog.Track("account_reg: %s %s %s", typ, str, pwd)

	if !format.CheckPasswd(pwd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if !format.CheckBindValue(typ, str) {
		ack.WriteUInt16(err.BindInfo_format_err)
	} else {
		errcode, _ := NewAccountInDB(pwd, typ, str)
		ack.WriteUInt16(errcode)
	}
}
func Rpc_center_reg_check(req, ack *common.NetPack) {
	str := req.ReadString()
	pwd := req.ReadString()
	typ := req.ReadString() //email、name、phone
	sign.Decode(&str, &pwd)

	if !format.CheckPasswd(pwd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if !format.CheckBindValue(typ, str) {
		ack.WriteUInt16(err.BindInfo_format_err)
	} else if GetAccountByBindInfo(typ, str) != nil {
		ack.WriteUInt16(err.Account_repeat)
	} else {
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_center_reg_if(req, ack *common.NetPack) {
	str := req.ReadString()
	typ := req.ReadString() //email、name、phone
	sign.Decode(&str)

	if GetAccountByBindInfo(typ, str) == nil {
		ack.WriteUInt16(err.Account_not_found)
	} else {
		ack.WriteUInt16(err.Success)
	}
}
func Rpc_center_account_reg_force(req, ack *common.NetPack) {
	uuid := req.ReadString()
	if sign.Decode(&uuid); uuid == "" {
		ack.WriteUInt16(err.BindInfo_format_err)
	} else {
		errcode, _ := NewAccountInDB("", "name", uuid)
		ack.WriteUInt16(errcode)
	}
}
func Rpc_center_change_password(req, ack *common.NetPack) {
	str := req.ReadString()
	oldpwd := req.ReadString()
	newpwd := req.ReadString()
	sign.Decode(&oldpwd, &newpwd)

	if !format.CheckPasswd(newpwd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if e, p := GetAccount(str, oldpwd); p == nil {
		ack.WriteUInt16(e)
	} else {
		p.SetPasswd(newpwd)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{
			"password": p.Password}})
		ack.WriteUInt16(err.Success)
		CacheDel(p)
	}
}
func Rpc_center_create_visitor(req, ack *common.NetPack) {
	name := fmt.Sprintf("%s%d%s",
		random.String(3),
		dbmgo.GetNextIncId("VisitorId"),
		random.String(3))

	if _, p := NewAccountInDB("", "name", name); p == nil {
		name = "" //failed
	}
	ack.WriteString(name)
}

// ------------------------------------------------------------
// 记录于账号上面的游戏信息，一套账号系统可关联多个游戏
func Rpc_center_set_game_route(req, ack *common.NetPack) {
	accountId := req.ReadUInt32()
	gameName := req.ReadString()
	loginId := req.ReadInt()
	gameId := req.ReadInt()
	if p := GetAccountById(accountId); p == nil {
		ack.WriteUInt16(err.Account_not_found)
	} else {
		v, ok := p.GameInfo[gameName]
		v.LoginSvrId = loginId
		v.GameSvrId = gameId
		if dbmgo.UpdateIdSync(KDBTable, accountId, bson.M{"$set": bson.M{
			fmt.Sprintf("gameinfo.%s", gameName): v}}) {
			p.GameInfo[gameName] = v
			ack.WriteUInt16(err.Success)
			if ok { //转服了，通知调用者删缓存
				p.writeLoginCacheKey(ack)
			}
			CacheDel(p) //先更db，再删缓存
		} else {
			ack.WriteUInt16(err.GameInfo_set_fail)
		}
	}
}
func Rpc_center_get_game_route(req, ack *common.NetPack) {
	accountId := req.ReadUInt32()
	gameName := req.ReadString()
	if p := GetAccountById(accountId); p != nil {
		if v, ok := p.GameInfo[gameName]; ok {
			ack.WriteInt(v.LoginSvrId)
			ack.WriteInt(v.GameSvrId)
		}
	}
}
func Rpc_center_set_game_json(req, ack *common.NetPack) { //TODO:zhoumf:换accountId + pwd
	str := req.ReadString()
	typ := req.ReadString() //email、name、phone
	gameName := req.ReadString()
	json := req.ReadString()
	sign.Decode(&str)

	if p := GetAccountByBindInfo(typ, str); p == nil {
		ack.WriteUInt16(err.Account_not_found)
	} else {
		v := p.GameInfo[gameName]
		v.JsonData = json
		if dbmgo.UpdateIdSync(KDBTable, p.AccountID, bson.M{"$set": bson.M{
			fmt.Sprintf("gameinfo.%s", gameName): v}}) {
			p.GameInfo[gameName] = v
			ack.WriteUInt16(err.Success)
			CacheDel(p) //先更db，再删缓存
		} else {
			ack.WriteUInt16(err.GameInfo_set_fail)
		}
	}
}
func Rpc_center_get_game_json(req, ack *common.NetPack) {
	str := req.ReadString()
	typ := req.ReadString() //email、name、phone
	gameName := req.ReadString()
	sign.Decode(&str)

	if p := GetAccountByBindInfo(typ, str); p != nil {
		if v, ok := p.GameInfo[gameName]; ok {
			ack.WriteString(v.JsonData)
		}
	}
}

// ------------------------------------------------------------
// 玩家在哪个大区登录的
func Rpc_center_player_login_addr_2(req, ack *common.NetPack) { //TODO:zhoumf:待删除
	gameName := req.ReadString()
	str := req.ReadString()

	var p *TAccount
	if p = GetAccountByBindInfo("email", str); p == nil {
		if p = GetAccountByBindInfo("name", str); p == nil {
		}
	}
	if p == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		p.WriteLoginAddr(gameName, ack)
	}
}
func Rpc_center_player_login_addr(req, ack *common.NetPack) {
	str := req.ReadString()
	typ := req.ReadString() //email、name、phone
	gameName := req.ReadString()
	sign.Decode(&str)

	if p := GetAccountByBindInfo(typ, str); p == nil {
		ack.WriteUInt16(err.Account_not_found)
	} else {
		p.WriteLoginAddr(gameName, ack)
	}
}
func (self *TAccount) WriteLoginAddr(gameName string, ack *common.NetPack) {
	if info, ok := self.GameInfo[gameName]; !ok {
		ack.WriteUInt16(err.Not_found)
	} else if p := meta.GetMeta("login", info.LoginSvrId); p == nil {
		ack.WriteUInt16(err.Svr_not_working) //玩家有对应的登录服，但该服未启动
		/*
			新center只加到华北
			华南玩家，客户端删包重装，会问询自己哪个大区的
			正好，他测速跑到华北区问
			又正好，他被路由到了新center节点
			新center没有华南信息，err.Svr_not_working
		*/
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

// ------------------------------------------------------------
// 登录服缓存了部分账号信息，用以加速登录
func (self *TAccount) refreshLoginCache() {
	go meta.G_Metas.Range(func(k, v interface{}) bool {
		if p := v.(*meta.Meta); p.Module == "login" && !p.IsClosed {
			http.CallRpc(http.Addr(p.OutIP, p.Port()), enum.Rpc_login_del_account_cache,
				func(buf *common.NetPack) {
					self.writeLoginCacheKey(buf)
				}, nil)
		}
		return true
	})
}
func (self *TAccount) writeLoginCacheKey(buf *common.NetPack) {
	buf.WriteByte(byte(len(self.BindInfo)))
	for _, v := range self.BindInfo {
		buf.WriteString(v)
	}
}
