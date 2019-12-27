/***********************************************************************
* @ 凡修改账号信息，先更db，再删缓存

* @ 过期策略+内存淘汰
	· 定期删除+惰性删除
		· 每100ms随机抽取进行检查，有过期Key删除
		· 获取key时，若此key过期了就删除
	· 当内存不足以容纳新写入数据时，移除最近最少使用的 Key
		# maxmemory-policy allkeys-lru

* @ bug 缓存不一致
	· 各center节点均有缓存，玩家多次操作，被路由到不同节点
	· 比如：邮箱登录 -> center1 -> 手机登录 -> center2 -> 改密码 -> 再邮箱登录 -> 失败

* @ author zhoumf
* @ date 2019-12-17
***********************************************************************/
package account

import (
	"common"
	"dbmgo"
	"github.com/go-redis/redis"
	"gopkg.in/mgo.v2/bson"
	"netConfig/meta"
	"time"
)

var g_redis *redis.Client

func Init() {
	// 跟db在同个节点，无需redis，mongo自带缓存的
	if meta.GetMeta("db_account", meta.G_Local.SvrID).IP != meta.G_Local.IP {
		g_redis = redis.NewClient(&redis.Options{
			Addr:     ":6379",
			Password: "",
		})
		if _, e := g_redis.Ping().Result(); e != nil {
			panic(e)
		}
	}
}

func CacheAdd(key string, p *TAccount) {
	if p.Name != "" && p.BindInfo["name"] == "" { //FIXME：初版账号系统的遗祸~囧
		p.BindInfo["name"] = p.Name
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"bindinfo.name": p.Name}})
	}
	if g_redis != nil {
		buf, _ := common.T2B(p)
		g_redis.Set(key, buf, 24*time.Hour)
	}
}
func CacheGet(key string) *TAccount {
	if g_redis != nil {
		if ret, e := g_redis.Get(key).Result(); e == nil {
			var v TAccount
			if common.B2T(common.S2B(ret), &v) == nil {
				return &v
			}
		}
	}
	return nil
}
func CacheDel(p *TAccount) {
	if g_redis != nil {
		var keys []string
		for k, v := range p.BindInfo {
			keys = append(keys, k+v)
		}
		g_redis.Del(keys...)
	}
}
