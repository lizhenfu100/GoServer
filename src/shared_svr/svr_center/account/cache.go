/***********************************************************************
* @ 凡修改账号信息，先更db，再删缓存

* @ 过期策略+内存淘汰
	· 定期删除+惰性删除
		· 每100ms随机抽取进行检查，有过期Key删除
		· 获取key时，若此key过期了就删除
	· 当内存不足以容纳新写入数据时，移除最近最少使用的 Key
		# maxmemory-policy allkeys-lru

* @ 主center直连mongodb，所有都连同个redis缓存
	· redis部署在分流节点机器上
	· 删缓存时，各节点就都删了，保证数据只一份

* @ bug 多redis的缓存不一致
	· 各center节点均有缓存，玩家多次操作，被路由到不同节点
	· 比如：邮箱登录 -> center1 -> 手机登录 -> center2 -> 改密码 -> 再邮箱登录 -> 失败

* @ author zhoumf
* @ date 2019-12-17
***********************************************************************/
package account

import (
	"common"
	"dbmgo"
	"encoding/binary"
	"github.com/go-redis/redis"
	"gopkg.in/mgo.v2/bson"
	"netConfig/meta"
	"time"
)

var (
	g_redis    *redis.Client
	_use_cache = true
)

func Init() {
	g_redis = redis.NewClient(&redis.Options{Addr: meta.GetMeta("redis", 0).IP + ":6379"})
	if _, e := g_redis.Ping().Result(); e != nil {
		panic(e)
	}
	// 跟db在同个节点，无需redis，mongo自带缓存的
	if meta.GetMeta("db_center", meta.G_Local.SvrID).IP == meta.G_Local.IP {
		_use_cache = false
	}
}
func CacheAdd(p *TAccount) {
	if p.Name != "" && p.BindInfo["name"] == "" { //TODO:待删除
		p.BindInfo["name"] = p.Name
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"bindinfo.name": p.Name}})
	}
	if _use_cache {
		buf, _ := common.T2B(p)
		aid := _i2s(p.AccountID)
		g_redis.Set(aid, buf, 24*time.Hour)
		for _, v := range p.BindInfo {
			g_redis.Set(v, aid, 24*time.Hour)
		}
	}
}
func CacheGet(key string) *TAccount {
	if _use_cache {
		if aid, e := g_redis.Get(key).Result(); e == nil {
			if b, e := g_redis.Get(aid).Result(); e == nil {
				var v TAccount
				if common.B2T(common.S2B(b), &v) == nil {
					return &v
				}
			}
		}
	}
	return nil
}
func CacheDel(p *TAccount) {
	keys := []string{_i2s(p.AccountID)}
	for _, v := range p.BindInfo {
		keys = append(keys, v)
	}
	g_redis.Del(keys...)
}

func _i2s(id uint32) string {
	aid := make([]byte, 4)
	binary.LittleEndian.PutUint32(aid, id)
	return common.B2S(aid)
}
