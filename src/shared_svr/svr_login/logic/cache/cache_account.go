/***********************************************************************
* @ redis 缓存账号+密码，减轻center压力
* @ brief
	1、凡是涉及修改账号信息的，先更db，再删缓存
	2、数据格式参见：Rpc_center_account_login

* @ 过期策略、内存淘汰
	· 定期删除+惰性删除
		· 每100ms随机抽取进行检查，有过期Key删除
		· 获取key时，若此key过期了就删除
	· 当内存不足以容纳新写入数据时，在键空间中，移除最近最少使用的 Key
		# maxmemory-policy allkeys-lru

* @ author zhoumf
* @ date 2019-12-9
***********************************************************************/
package cache

import (
	"common"
	"common/std/sign"
	"gamelog"
	"generate_out/err"
	"github.com/go-redis/redis"
	"shared_svr/svr_center/account/gameInfo"
)

var g_client = redis.NewClient(&redis.Options{
	Addr:     ":6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

type TCache struct {
	gameInfo.TAccountClient
	LoginSvrId int
	GameSvrId  int
	Password   string
}

// 逻辑同 Rpc_center_account_login
func AccountLogin(req, ack *common.NetPack) (ok bool, account, pwd string) {
	oldPos := req.ReadPos //临时读取
	req.ReadString()      //gameName
	account = req.ReadString()
	pwd = req.ReadString()
	sign.Decode(&account, &pwd)
	req.ReadPos = oldPos
	if ret, e := g_client.Get(account).Result(); e == nil {
		var v TCache
		if common.B2T(common.S2B(ret), &v) == nil && v.Password == pwd {
			ok = true
			ack.WriteUInt16(err.Success)
			v.DataToBuf(ack)
			ack.WriteInt(v.LoginSvrId)
			ack.WriteInt(v.GameSvrId)
		}
	}
	gamelog.Debug("Cache Login: %v", ok)
	return
}
func Add(account, pwd string, ack *common.NetPack) {
	oldPos := ack.ReadPos //临时读取
	if ack.ReadUInt16() == err.Success {
		var v TCache
		v.BufToData(ack)
		v.LoginSvrId = ack.ReadInt()
		v.GameSvrId = ack.ReadInt()
		v.Password = pwd
		//有效的登录信息，才缓存
		if v.LoginSvrId > 0 && v.GameSvrId > 0 {
			buf, _ := common.T2B(&v)
			g_client.Set(account, buf, 0)
		}
	}
	ack.ReadPos = oldPos
}
func Rpc_login_del_account_cache(req, ack *common.NetPack) {
	for cnt, i := req.ReadByte(), byte(0); i < cnt; i++ {
		g_client.Del(req.ReadString())
	}
}
