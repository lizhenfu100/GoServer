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
	name string //<id, score>
	show string //<id, showInfo>
}

func (p *TScore) Set(member string, score float64, pShow interface{}) {
	g_redis.ZAdd(p.name, redis.Z{ //会更新已有成员的分数
		Score:  score,
		Member: member,
	})
	buf, _ := common.T2B(pShow)
	g_redis.HSet(p.show, member, buf)
}
func (p *TScore) SetShow(member string, pShow interface{}) {
	buf, _ := common.T2B(pShow)
	g_redis.HSet(p.show, member, buf)
}
func (p *TScore) GetShow(member string) []byte { //common.B2T
	if b, e := g_redis.HGet(p.show, member).Result(); e == nil {
		return common.S2B(b)
	}
	return nil
}
func (p *TScore) GetScore(member string) float64 {
	return g_redis.ZScore(p.name, member).Val()
}
func (p *TScore) AddScore(member string, score float64) float64 {
	return g_redis.ZIncrBy(p.name, score, member).Val()
}
func (p *TScore) Rank(member string) int {
	if v, e := g_redis.ZRevRank(p.name, member).Result(); e == nil {
		return int(v)
	}
	return -1
}
func (p *TScore) Del(member string) {
	g_redis.ZRem(p.name, member)
	g_redis.HDel(p.show, member)
}
func (p *TScore) Clear() { g_redis.Del(p.name, p.show) }
