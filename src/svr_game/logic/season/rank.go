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
	"common/std/rank"
	"svr_game/conf"
)

var (
	KRankNeedScore int //入排行榜所需积分
	g_ranker       rank.TRanker
	g_rank_map     map[uint32]*RankItem //<pid, >
)

// ------------------------------------------------------------
// -- 排行榜 rank.IRankItem
type RankItem struct {
	Rank uint8 `bson:"_id"` //0无效，1起始
	//显示信息，玩家数据的拷贝
	Score    int
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
	kLen := len(conf.Const.Season_Score)
	KRankNeedScore = conf.Const.Season_Score[kLen-1]
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
