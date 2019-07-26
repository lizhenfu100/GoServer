/***********************************************************************
* @ 赛季积分
* @ brief
	1、赛季结束后清空数据（玩家登录时检查）

* @ 排名机制
                  到下一级积分 		总共场次		 差值场次
    0 Unranked		0
	1 Bronze	   	250        			 1		   1
	2 Silver	   	500       			10		   9
	3 Gold	   		1500				30		  20
	4 Platinum   	5000			   100		  70
	5 Diamond	   	10000              200		 100
	6 Master	   	20000        	   400		 200
	7 Grandmaster  	只有排名, 前100，按积分排名

* @ 获得积分规则:
	正常完成基础分+50。正常完成定义为打出伤害，存活超过30秒就有。中途主动退出为0分
	胜利+100
	击杀第一+30，第二+20，第三+15，前50%+10
	辅助击杀第一+30，第二+20，第三+10，前50%+5
	拉起队友一次+5，上限15

* @ author zhoumf
* @ date 2018-5-8
***********************************************************************/
package player

import (
	"common"
	"dbmgo"
	"svr_game/conf"
	"svr_game/logic/season"
)

const kDBSeason = "season"

type TSeasonModule struct {
	PlayerID  uint32 `bson:"_id"`
	TimeBegin int64  //赛季开启时刻，仅用于识别“离线跨赛季”
	Score     int    //赛季积分

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
func (self *TSeasonModule) AddScore(diff int) {
	if self.Score += diff; self.Score < 0 {
		self.Score = 0
	}
	if self.Score >= season.KRankNeedScore {
		if p := self._RankItem(); p.OnValueChange() {
			season.AddRankItem(p)
		}
	}
}
func (self *TSeasonModule) calcScore(
	isWin bool,
	killCnt, //击杀数
	assistCnt, //助攻数
	reviveCnt uint8) int { //拉队友次数

	kConf := &conf.Const
	ret := kConf.Score_Normal
	if isWin {
		ret = kConf.Score_Win
	}
	for i, kLen := 0, len(kConf.Score_Kill); i < kLen; i++ {
		if killCnt > 0 {
			killCnt--
			ret += int(kConf.Score_Kill[i])
		}
	}
	for i, kLen := 0, len(kConf.Score_Assist); i < kLen; i++ {
		if assistCnt > 0 {
			assistCnt--
			ret += int(kConf.Score_Assist[i])
		}
	}
	if n := reviveCnt * kConf.Score_Revive; n < kConf.Score_Revive_Max {
		ret += int(n)
	} else {
		ret += int(kConf.Score_Revive_Max)
	}
	return ret
}
func (self *TSeasonModule) calcLv() uint8 {
	kConf := &conf.Const
	kLen := uint8(len(kConf.Season_Score))
	for i := uint8(0); i < kLen; i++ {
		if self.Score < kConf.Season_Score[i] {
			return i
		}
	}
	return kLen
}

func (self *TSeasonModule) GetRank() uint8 {
	if v := season.GetRankItem(self.PlayerID); v != nil {
		return v.Rank
	} else {
		return 0
	}
}
func (self *TSeasonModule) Clear() {
	self.TimeBegin = season.G_Args.TimeBgein
	self.Score = 0
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
	ack.WriteInt(this.season.Score)
	ack.WriteUInt8(this.season.calcLv())
	ack.WriteUInt8(this.season.GetRank()) //仅最高档有排名，其余为0
}
