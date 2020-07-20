package conf

import "common/std"

//go:generate D:\server\bin\gen_conf.exe conf *csv
type csv struct {
	LoginSvrId int //所在区服

	DailySignInReward [][]std.IntPair //每日签到奖励

	// 赛季
	Season_Begin_Month []int //开始月份
	Season_Score       []int //赛季，各档次积分门槛
	Score_Normal       int   //正常完成基础分。正常完成定义为打出伤害，存活超过30秒就有。中途主动退出为0分
	Score_Win          int
	Score_Kill         []uint8  //击杀得分，第一次、第二次、第三次...
	Score_Assist       []uint8  //助攻得分
	Score_Revive       uint8    //拉起队友得分
	Score_Revive_Max   uint8    //拉队友得分上限
	Score_Once_Max     uint8    //单场得分上限
	HeroExp_Win        uint8    //胜利英雄经验
	HeroExp_Fail       uint8    //失败英雄经验
	Hero_LvUp          []uint16 //升级经验分布
	Gun_LvUp           []uint16 //升级经验分布
}
