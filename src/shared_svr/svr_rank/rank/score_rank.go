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
	Name string //<id, score>
	Show string //<id, showInfo>
}

func (p *TScore) Set(member string, score float64, pShow interface{}) {
	g_redis.ZAdd(p.Name, redis.Z{ //会更新已有成员的分数
		Score:  score,
		Member: member,
	})
	buf, _ := common.T2B(pShow)
	g_redis.HSet(p.Show, member, buf)
}
func (p *TScore) SetShow(member string, pShow interface{}) {
	buf, _ := common.T2B(pShow)
	g_redis.HSet(p.Show, member, buf)
}
func (p *TScore) GetShow(member string) []byte { //common.B2T
	if b, e := g_redis.HGet(p.Show, member).Result(); e == nil {
		return common.S2B(b)
	}
	return nil
}
func (p *TScore) GetScore(member string) float64 {
	return g_redis.ZScore(p.Name, member).Val()
}
func (p *TScore) AddScore(member string, score float64) float64 {
	return g_redis.ZIncrBy(p.Name, score, member).Val()
}
func (p *TScore) Rank(member string) int {
	if v, e := g_redis.ZRevRank(p.Name, member).Result(); e == nil {
		return int(v)
	}
	return -1
}
func (p *TScore) Del(member string) {
	g_redis.ZRem(p.Name, member)
	g_redis.HDel(p.Show, member)
}
func (p *TScore) Clear() { g_redis.Del(p.Name, p.Show) }
