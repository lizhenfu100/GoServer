/***********************************************************************
* @ 赛季积分
* @ brief
	1、赛季结束后清空数据（玩家登录时检查）

* @ 排名机制
				  分档  积分
	0 Bronze	   5 * 100
	1 Silver	   5 * 100
	2 Gold	       5 * 100
	3 Platinum     5 * 100
	4 Diamond	   5 * 100
	5 Master	   5 * 100
	6 GrandMaster  只有排名, 前100

* @ author zhoumf
* @ date 2018-5-8
***********************************************************************/
package player

import (
	"common"
	"common/std/random"
	"dbmgo"
	"math/rand"
	"svr_game/conf"
	"svr_game/player/season"
)

const kDBSeason = "season"

type TSeasonModule struct {
	PlayerID  uint32 `bson:"_id"`
	Score     uint16 //赛季积分
	Level     uint8  //所处档次
	SecondLv  uint8  //每个档的小级别
	TimeBegin int64  //赛季开启时刻，仅用于识别“离线跨赛季”

	owner     *TPlayer
	winStreak byte //连胜纪录，影响赛季得分
}

// ------------------------------------------------------------
// -- 框架接口
func (self *TSeasonModule) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.Insert(kDBSeason, self)
	self._InitTempData(player)
}
func (self *TSeasonModule) LoadFromDB(player *TPlayer) {
	if ok, _ := dbmgo.Find(kDBSeason, "_id", player.PlayerID, self); !ok {
		self.InitAndInsert(player)
	}
	self._InitTempData(player)
}
func (self *TSeasonModule) WriteToDB() { dbmgo.UpdateId(kDBSeason, self.PlayerID, self) }
func (self *TSeasonModule) OnLogin() {
	if self.TimeBegin != season.G_Args.TimeBgein {
		self.Clear()
	}
}
func (self *TSeasonModule) OnLogout() {
}
func (self *TSeasonModule) _InitTempData(player *TPlayer) {
	self.owner = player
}

// ------------------------------------------------------------
// -- API
func InitSeasonDB() {
	season.InitRankList()
	//赛季时间
	if !season.G_Args.ReadDB() {
		season.G_Args.TimeBgein = season.GetBeginTime()
		season.G_Args.InsertDB()
	}
}

func (self *TSeasonModule) AddScore(diff int) {
	if diff > conf.Const.Score_Once_Max {
		diff = conf.Const.Score_Once_Max
	}
	//1、变更积分
	score := 0
	if score = int(self.Score) + diff; score < 0 {
		score = 0
	}
	self.Score = uint16(score)
	//2、刷新赛季档次
	if self.Score >= season.KRankNeedScore {
		self.Level = conf.Const.Season_Level_Max
		self.SecondLv = 0
		if p := self._RankItem(); p.OnValueChange() {
			season.AddRankItem(p)
		}
	} else {
		confLevelScore := conf.Const.Season_Second_Level_Cnt * conf.Const.Season_Second_Level_Score
		self.Level = uint8(self.Score / confLevelScore)
		leftScore := self.Score % confLevelScore
		self.SecondLv = 1 + uint8(leftScore/conf.Const.Season_Second_Level_Score)
	}
}
func (self *TSeasonModule) calcScore(isWin bool, rank float32) int {
	if isWin == false {
		score := random.Between(conf.Const.Score_OneGame[0], conf.Const.Score_OneGame[1])
		if self.winStreak > 0 {
			ratio := rand.Float32() + 1 //连胜系数
			score = int(float32(score) * ratio)
		}
		return score
	} else {
		confRank := conf.Const.Score_Take_Off[self.Level][0]
		confScore := int(conf.Const.Score_Take_Off[self.Level][1])
		if rank > confRank {
			return -random.Between(1, confScore)
		} else {
			return 0
		}
	}
}

func (self *TSeasonModule) GetRank() uint8 {
	if v := season.GetRankItem(self.PlayerID); v != nil {
		return v.Rank
	} else {
		return 0
	}
}
func (self *TSeasonModule) Clear() {
	self.Score = 0
	self.Level = 0
	self.SecondLv = 0
	self.TimeBegin = season.G_Args.TimeBgein
}

// ------------------------------------------------------------
// -- 排行榜
func (self *TSeasonModule) _RankItem() *season.RankItem {
	if v := season.GetRankItem(self.PlayerID); v != nil {
		return v
	} else {
		return &season.RankItem{
			0,
			self.Score,
			self.PlayerID,
			self.owner.Name,
		}
	}
}

// ------------------------------------------------------------
// -- rpc
func Rpc_game_season_info(req, ack *common.NetPack, this *TPlayer) { //客户端界面查看赛季信息
	if this.season.TimeBegin != season.G_Args.TimeBgein {
		this.season.Clear()
	}
	ack.WriteUInt16(this.season.Score)
	ack.WriteUInt8(this.season.Level)
	ack.WriteUInt8(this.season.SecondLv)
	ack.WriteUInt8(this.season.GetRank()) //仅最高档有排名，其余为0
}
