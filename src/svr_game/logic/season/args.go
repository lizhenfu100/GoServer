/***********************************************************************
* @ 赛季排名机制
* @ brief
                  到下一级积分 		总共场次		 差值场次
    0 Unranked		0
	1 Bronze	   	250        			 1		   1
	2 Silver	   	500       			10		   9
	3 Gold	   		1500				30		  20
	4 Platinum   	5000			   100		  70
	5 Diamond	   	10000              200		 100
	6 Master	   	20000        	   400		 200
	7 Grandmaster  	只有排名, 前100，按积分排名

* @ author zhoumf
* @ date 2018-5-8
***********************************************************************/
package season

import (
	"dbmgo"
	"shared_svr/svr_rank/rank"
	"sort"
	"svr_game/conf"
	"time"
)

var (
	G_Args = stArgs{Key: "season"}

	KRankNeedScore int //入排行榜所需积分

	G_Rank = rank.TScore{"season", "season_"}
)

type stArgs struct {
	Key       string `bson:"_id"`
	TimeBgein int64  //服务器当前赛季的开启时刻
}

func (self *stArgs) ReadDB() bool {
	ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", self.Key, self)
	return ok
}
func (self *stArgs) UpdateDB() { dbmgo.UpdateId(dbmgo.KTableArgs, self.Key, self) }
func (self *stArgs) InsertDB() { dbmgo.Insert(dbmgo.KTableArgs, self) }

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

	if !G_Args.ReadDB() { //赛季时间
		G_Args.TimeBgein = GetBeginTime()
		G_Args.InsertDB()
	}
}

func GetBeginTime() int64 { //本赛季开启时刻
	now := time.Now()
	month := int(now.Month())
	csv := conf.Csv()
	length := len(csv.Season_Begin_Month)
	idx := sort.Search(length, func(i int) bool {
		return csv.Season_Begin_Month[i] > month
	})
	if idx == 0 {
		idx = length
	}
	month = csv.Season_Begin_Month[idx-1]
	return time.Date(now.Year(), time.Month(month), 1, 0, 0, 0, 0, now.Location()).Unix()
}
func OnEnterNextDay() {
	if t := GetBeginTime(); t > G_Args.TimeBgein {
		G_Args.TimeBgein = t
		G_Args.UpdateDB()
		G_Rank.Clear()
	}
}
