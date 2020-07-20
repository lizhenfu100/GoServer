package rank

import (
	"common"
	"github.com/go-redis/redis"
)

var g_redis = redis.NewClient(&redis.Options{Addr: ":6379"})

func init() {
	if _, e := g_redis.Ping().Result(); e != nil {
		panic(e)
	}
}

type TScore struct {
	Name string //分数表 <key, score>
	Show string //显示表 <key, showInfo>
}

func (p *TScore) Set(key string, score float64, pInfo interface{}) {
	g_redis.ZAdd(p.Name, redis.Z{ //会更新已有成员的分数
		Score:  score,
		Member: key,
	})
	buf, _ := common.T2B(pInfo)
	g_redis.HSet(p.Show, key, buf)
}
func (p *TScore) SetInfo(key string, pInfo interface{}) {
	buf, _ := common.T2B(pInfo)
	g_redis.HSet(p.Show, key, buf)
}
func (p *TScore) GetInfo(key string) []byte { //common.B2T
	if b, e := g_redis.HGet(p.Show, key).Result(); e == nil {
		return common.S2B(b)
	}
	return nil
}
func (p *TScore) GetScore(key string) float64 {
	return g_redis.ZScore(p.Name, key).Val()
}
func (p *TScore) AddScore(key string, score float64) float64 {
	return g_redis.ZIncrBy(p.Name, score, key).Val()
}
func (p *TScore) Rank(key string) int {
	if v, e := g_redis.ZRevRank(p.Name, key).Result(); e == nil {
		return int(v)
	}
	return -1
}
func (p *TScore) Del(key string) {
	g_redis.ZRem(p.Name, key)
	g_redis.HDel(p.Show, key)
}
func (p *TScore) Clear() { g_redis.Del(p.Name, p.Show) }
