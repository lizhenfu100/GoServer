/***********************************************************************
* @ 赛季排名机制
* @ brief
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
package season

import (
	"common/std/rank"
	"svr_game/conf"
)

var (
	KRankNeedScore uint16 //入排行榜所需积分
	g_ranker       rank.TRanker
	g_rank_map     map[uint32]*RankItem //<pid, >
)

// ------------------------------------------------------------
// -- 排行榜 rank.IRankItem
type RankItem struct {
	Rank uint8 `bson:"_id"`
	//显示信息，玩家数据的拷贝
	Score    uint16
	PlayerID uint32
	Name     string
}

func (self *RankItem) GetRank() uint { return uint(self.Rank) }
func (self *RankItem) SetRank(i uint) {
	if self.Rank = uint8(i); i == 0 {
		delete(g_rank_map, self.PlayerID) //被挤出排行榜
	}
}
func (self *RankItem) Less(obj rank.IRankItem) bool { return self.Score < obj.(*RankItem).Score }
func (self *RankItem) OnValueChange() bool          { return g_ranker.OnValueChange(self) }

// ------------------------------------------------------------
//
func InitRankList() {
	KRankNeedScore = uint16(conf.Const.Season_Level_Max) *
		conf.Const.Season_Second_Level_Cnt *
		conf.Const.Season_Second_Level_Score
	//初始化排行榜
	var list []RankItem
	g_ranker.Init("SeasonRank", 100, &list)
	g_rank_map = make(map[uint32]*RankItem, 100)
	for i := 0; i < len(list); i++ {
		ptr := &list[i]
		g_rank_map[ptr.PlayerID] = ptr
		g_ranker.InsertToIndex(ptr.GetRank(), ptr)
	}
}
func GetRankItem(pid uint32) *RankItem {
	if v, ok := g_rank_map[pid]; ok {
		return v
	}
	return nil
}
func AddRankItem(p *RankItem) { g_rank_map[p.PlayerID] = p }
