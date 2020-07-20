/***********************************************************************
* @ 账号管理 【 center的rpc都须返回错误码 】

* @ 目前，所有center同质化的，连一个db
	· client尽量少调center，许多接口结果可缓存

* @ 若注册数量到亿级，必须分库，“哈希id”或“id分段”至多个db TODO:optimize
	· 哈希：每次扩容须迁移部分账号
	· 分段：新玩家集中，单点压力大 …… 倾向于JumpHash方式
	· 玩家分别用 “账号名、邮箱、手机号” 登录，如何快速定位节点？
		· <字符串,ID>映射，哈希字符串，同样存到多个db
		· 须保证注册、改名、改邮箱...事务性修改映射表~囧~研究下saga模式
		· 先查映射表，再查账号数据

* @ 作业队列
	· 记录一条任务
	· 幂等的操作两方，完成后修改任务状态为Done
	· 只要任务不是Done，重试

* @ 二阶段提交
	· 在transactions集合中插入转账信息
		· db.transactions.insert({src:"A", dst:"B", value:100, state:"init"})
		· state：init，pending，committed，done，canceling，canceled
	1、找init事务
		· db.transactions.findOne({state: "init"})
	2、state更新为pending
		· db.transactions.update({_id:t._id, state:"init"}, {$set:{state:"pending"}})
	3、账户应用事务，须记录已应用的事务，防重复执行
		· db.accounts.update({_id:t.src, pendingTransactions:{$ne:t._id}}, {
				$inc:{money: -t.value},
				$push:{pendingTransactions:t._id}})
		· db.accounts.update({_id:t.dst, pendingTransactions:{$ne:t._id}}, {
				$inc:{money: t.value},
				$push:{pendingTransactions:t._id}})
	4、state更新为committed
		· db.transactions.update({_id:t._id, state:"pending"}, {$set: {state:"committed"}})
	5、账户移除事务
		· db.accounts.update({_id: t.src},{ $pull:{pendingTransactions:t._id}})
		· db.accounts.update({_id: t.dst},{ $pull:{pendingTransactions:t._id}})
	6、state更新为done
		· db.transactions.update({_id:t._id, state:"committed"}, {$set:{state:"done"}})
	· 故障恢复
		· 找出pending事务，从步骤3开始恢复
		· 找出committed事务，从步骤5开始恢复

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
	"common/std/sign"
	"crypto/md5"
	"dbmgo"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"gopkg.in/mgo.v2/bson"
	"netConfig/meta"
	"nets/http"
	"shared_svr/svr_center/account/gameInfo"
	"sync/atomic"
	"time"
)

const KDBTable = "Account"

type TAccount struct {
	AccountID    uint32 `bson:"_id"`
	Name         string //TODO:待删除
	Password     string //FIXME:可以StrHash后存成uint32，省不少字节
	CreateTime   int64
	LoginTime    int64
	IsValidEmail uint8
	BindInfo     map[string]string //email、phone、name
	GameInfo     map[string]gameInfo.TGameInfo
}

func _NewAccount() *TAccount {
	self := new(TAccount)
	self.BindInfo = make(map[string]string, 2)
	self.GameInfo = make(map[string]gameInfo.TGameInfo)
	return self
}
func (self *TAccount) CheckPasswd(passwd string) bool {
	return self.Password == fmt.Sprintf("%x", md5.Sum(common.S2B(passwd)))
}
func (self *TAccount) SetPasswd(passwd string) {
	self.Password = fmt.Sprintf("%x", md5.Sum(common.S2B(passwd)))
}

// ------------------------------------------------------------
// 注册、登录
func Rpc_center_account_login(req, ack *common.NetPack, _ common.Conn) {
	gameName := req.ReadString()
	str := req.ReadString()
	pwd := req.ReadString()
	errcode, p := GetAccount(str, pwd)
	if ack.WriteUInt16(errcode); p != nil {
		timeNow := time.Now().Unix()
		atomic.StoreInt64(&p.LoginTime, timeNow)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"logintime": timeNow}})
		(&gameInfo.TAccountClient{ //1、先回复，Client账号信息
			p.AccountID,
			p.IsValidEmail,
			p.BindInfo,
		}).DataToBuf(ack)
		if v, ok := p.GameInfo[gameName]; ok { //2、附带的游戏数据，可能有的游戏空的
			ack.WriteInt(v.LoginSvrId)
			ack.WriteInt(v.GameSvrId)
		}
	}
}
func Rpc_center_account_reg(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	pwd := req.ReadString()
	typ := req.ReadString() //email、name、phone
	sign.Decode(&str, &pwd)
	if !format.CheckPasswd(pwd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if !format.CheckBindValue(typ, str) {
		ack.WriteUInt16(err.BindInfo_format_err)
	} else {
		errcode, _ := NewAccountInDB(pwd, typ, str)
		ack.WriteUInt16(errcode)
	}
}
func Rpc_center_account_reg_force(req, ack *common.NetPack, _ common.Conn) { //TODO:待删除2020.7.2
	uuid := req.ReadString()
	if sign.Decode(&uuid); uuid == "" {
		ack.WriteUInt16(err.BindInfo_format_err)
	} else {
		e, _ := NewAccountInDB("", "name", uuid)
		ack.WriteUInt16(e)
	}
}
func Rpc_center_reg_if(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	typ := req.ReadString()
	sign.Decode(&str)
	e, _ := GetAccountByBindInfo(typ, str)
	if e == err.Not_found {
		e = err.Account_not_found
	}
	ack.WriteUInt16(e)
}
func Rpc_center_change_password2(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	typ := req.ReadString()
	oldpwd := req.ReadString()
	newpwd := req.ReadString()
	sign.Decode(&str, &oldpwd, &newpwd)
	if !format.CheckPasswd(newpwd) {
		ack.WriteUInt16(err.Passwd_format_err)
	} else if e, p := GetAccountByBindInfo(typ, str); p == nil {
		ack.WriteUInt16(e)
	} else if !p.CheckPasswd(oldpwd) {
		ack.WriteUInt16(err.Account_mismatch_passwd)
	} else {
		ack.WriteUInt16(err.Success)
		p.SetPasswd(newpwd)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{
			"password": p.Password}})
		CacheDel(p)
	}
}

// ------------------------------------------------------------
// 记录于账号上面的游戏信息，一套账号系统可关联多个游戏
func Rpc_center_set_game_route(req, ack *common.NetPack, _ common.Conn) {
	accountId := req.ReadUInt32()
	gameName := req.ReadString()
	loginId := req.ReadInt()
	gameId := req.ReadInt()
	if e, p := GetAccountById(accountId); p == nil {
		ack.WriteUInt16(e)
	} else {
		v, ok := p.GameInfo[gameName]
		v.LoginSvrId = loginId % common.KIdMod
		v.GameSvrId = gameId % common.KIdMod
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
func Rpc_center_set_game_json(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	typ := req.ReadString()
	gameName := req.ReadString()
	json := req.ReadString()
	sign.Decode(&str)
	if e, p := GetAccountByBindInfo(typ, str); p == nil {
		ack.WriteUInt16(e)
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
func Rpc_center_get_game_json(req, ack *common.NetPack, _ common.Conn) { //TODO:待删除
	str := req.ReadString()
	typ := req.ReadString()
	gameName := req.ReadString()
	sign.Decode(&str)
	if _, p := GetAccountByBindInfo(typ, str); p != nil {
		if v, ok := p.GameInfo[gameName]; ok {
			ack.WriteString(v.JsonData)
		}
	}
}

// ------------------------------------------------------------
// 玩家在哪个大区登录的
func Rpc_center_player_login_addr_2(req, ack *common.NetPack, _ common.Conn) { //TODO:待删除
	gameName := req.ReadString()
	str := req.ReadString()

	var p *TAccount
	if _, p = GetAccountByBindInfo("email", str); p == nil {
		if _, p = GetAccountByBindInfo("name", str); p == nil {
		}
	}
	if p == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		p.WriteLoginAddr(gameName, ack)
	}
}
func Rpc_center_player_login_addr(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	typ := req.ReadString() //email、name、phone
	gameName := req.ReadString()
	sign.Decode(&str)
	if e, p := GetAccountByBindInfo(typ, str); p == nil {
		ack.WriteUInt16(e)
	} else {
		p.WriteLoginAddr(gameName, ack)
	}
}
func (self *TAccount) WriteLoginAddr(gameName string, ack *common.NetPack) {
	errcode, gameSvrId := err.Success, 0
	var pLogin *meta.Meta
	if info, ok := self.GameInfo[gameName]; ok {
		gameSvrId = info.GameSvrId
		if pLogin = meta.GetMeta(gameName, info.LoginSvrId); pLogin == nil {
			errcode = err.Svr_not_working
			/*
				新center只加到华北
				华南玩家，客户端删包重装，会问询自己哪个大区的
				正好，他测速跑到华北区问
				又正好，他被路由到了新center节点
				新center没有华南信息，err.Svr_not_working
			*/
		}
	}
	if ack.WriteUInt16(errcode); pLogin != nil {
		ack.WriteString(pLogin.OutIP)
		ack.WriteUInt16(pLogin.HttpPort)
		// 向login查询game的地址，一并回给client
		http.CallRpc(http.Addr(pLogin.OutIP, pLogin.HttpPort), enum.Rpc_get_meta, func(buf *common.NetPack) {
			buf.WriteString("game")
			buf.WriteString(meta.G_Local.Version)
			buf.WriteByte(common.KeyShuntInt)
			buf.WriteInt(gameSvrId)
			buf.WriteUInt32(self.AccountID)
		}, func(recvBuf *common.NetPack) {
			if svrId := recvBuf.ReadInt(); svrId > 0 {
				ip := recvBuf.ReadString()
				port := recvBuf.ReadUInt16()
				ack.WriteString(ip)
				ack.WriteUInt16(port)
			}
		})
	}
}

// ------------------------------------------------------------
// 登录服缓存了部分账号信息，用以加速登录 ---- 遍历登录服，太挫了
func (self *TAccount) writeLoginCacheKey(buf *common.NetPack) {
	buf.WriteByte(byte(len(self.BindInfo)))
	for _, v := range self.BindInfo {
		buf.WriteString(v)
	}
}
