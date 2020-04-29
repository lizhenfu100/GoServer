package season

import (
	"dbmgo"
	"sort"
	"svr_game/conf"
	"time"
)

var G_Args = stArgs{Key: "season"}

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
// -- API
func InitDB() {
	InitRankList()
	//赛季时间
	if !G_Args.ReadDB() {
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
		g_rank.Clear()
		for k := range g_items {
			delete(g_items, k)
		}
	}
}
