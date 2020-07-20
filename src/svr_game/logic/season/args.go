package season

import (
	"dbmgo"
	"shared_svr/svr_rank/rank"
	"sort"
	"svr_game/conf"
	"time"
)

var (
	G_Args = Args{Key: "season"}
	G_Rank = rank.TScore{"season", "season_"}

	KRankNeedScore int //入排行榜所需积分
)

type Args struct {
	Key       string `bson:"_id"`
	TimeBgein int64  //服务器当前赛季的开启时刻
	Idx       uint8  //第几赛季
}

func (self *Args) ReadDB() bool {
	ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", self.Key, self)
	return ok
}
func (self *Args) UpdateDB() { dbmgo.UpdateId(dbmgo.KTableArgs, self.Key, self) }

// ------------------------------------------------------------
type RankInfo struct {
	Score float64
	Pid   string
	Name  string
}

// ------------------------------------------------------------
// -- API
func InitDB() {
	csv := conf.Csv()
	KRankNeedScore = csv.Season_Score[len(csv.Season_Score)-1]
	if t, i := GetBeginTime(); !G_Args.ReadDB() || t > G_Args.TimeBgein { //赛季时间
		G_Args.TimeBgein, G_Args.Idx = t, i
		dbmgo.UpsertId(dbmgo.KTableArgs, G_Args.Key, &G_Args)
	}
}
func GetBeginTime() (int64, uint8) { //本赛季开启时刻
	now := time.Now()
	month, csv := int(now.Month()), conf.Csv()
	length := len(csv.Season_Begin_Month)
	idx := sort.Search(length, func(i int) bool {
		return csv.Season_Begin_Month[i] > month
	})
	if idx == 0 {
		idx = length
	}
	idx--
	month = csv.Season_Begin_Month[idx]
	return time.Date(now.Year(), time.Month(month), 1, 0, 0, 0, 0, now.Location()).Unix(), uint8(idx)
}
func OnEnterNextDay() {
	if t, i := GetBeginTime(); t > G_Args.TimeBgein {
		G_Args.TimeBgein, G_Args.Idx = t, i
		G_Args.UpdateDB()
		G_Rank.Clear()
	}
}
