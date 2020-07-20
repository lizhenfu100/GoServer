/***********************************************************************
* @ redis 缓存账号+密码，减轻center压力
* @ brief
	1、凡修改账号信息，先更db，再删缓存
	2、数据格式参见：Rpc_center_account_login

* @ 过期策略+内存淘汰
	· 定期删除+惰性删除
		· 每100ms随机抽取进行检查，有过期Key删除
		· 获取key时，若此key过期了就删除
	· 当内存不足以容纳新写入数据时，移除最近最少使用的 Key
		# maxmemory-policy allkeys-lru

* @ bug 缓存不一致
	· 密码特殊处理过，不一致会去center查
	· 路由数据，迁服rpc回调里及时删缓存
	· 邮箱、手机是否验证，center广播失败，就只能等缓存失效了

* @ author zhoumf
* @ date 2019-12-9
***********************************************************************/
package cache

import (
	"common"
	"conf"
	"encoding/binary"
	"gamelog"
	"generate_out/err"
	"github.com/go-redis/redis"
	"shared_svr/svr_center/account/gameInfo"
	"time"
)

var g_redis = redis.NewClient(&redis.Options{Addr: ":6379"})

func init() {
	if _, e := g_redis.Ping().Result(); e != nil {
		panic(e)
	}
}

type TCache struct {
	gameInfo.TAccountClient
	LoginSvrId int
	GameSvrId  int
	Password   string
}

func Add(pwd string, ack *common.NetPack) {
	p, oldPos := &TCache{Password: pwd}, ack.ReadPos
	if ack.ReadUInt16() == err.Success {
		p.BufToData(ack)
		p.LoginSvrId = ack.ReadInt()
		p.GameSvrId = ack.ReadInt()
		if !conf.Auto_GameSvr || p.GameSvrId > 0 {
			buf, _ := common.T2B(p)
			aid := _i2s(p.AccountID)
			g_redis.Set(aid, buf, 48*time.Hour)
			for _, v := range p.BindInfo {
				g_redis.Set(v, aid, 48*time.Hour)
			}
		}
	}
	ack.ReadPos = oldPos //临时读取
}
func Get(account, pwd string, ack *common.NetPack) (ok bool) {
	if p := _get(account); p != nil && p.Password == pwd {
		ok = true
		ack.WriteUInt16(err.Success)
		p.DataToBuf(ack)
		ack.WriteInt(p.LoginSvrId)
		ack.WriteInt(p.GameSvrId)
	}
	gamelog.Track("Cache Login: %s %v", account, ok)
	return
}
func _get(account string) *TCache {
	if aid, e := g_redis.Get(account).Result(); e == nil {
		if b, e := g_redis.Get(aid).Result(); e == nil {
			var v TCache
			if common.B2T(common.S2B(b), &v) == nil {
				return &v
			}
		}
	}
	return nil
}
func Del(account string) {
	if p := _get(account); p != nil {
		keys := []string{_i2s(p.AccountID)}
		for _, v := range p.BindInfo {
			keys = append(keys, v)
		}
		g_redis.Del(keys...)
	}
}

func _i2s(id uint32) string {
	aid := make([]byte, 4)
	binary.LittleEndian.PutUint32(aid, id)
	return common.B2S(aid)
}

func IsExist(account string) bool {
	_, e := g_redis.Get(account).Result()
	return e == nil
}
func Rpc_login_del_account_cache(req, ack *common.NetPack, _ common.Conn) {
	cnt := req.ReadByte()
	keys := make([]string, cnt)
	for i := byte(0); i < cnt; i++ {
		keys[i] = req.ReadString()
	}
	g_redis.Del(keys...)
}
