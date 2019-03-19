package conf

var Const struct {
	// 赛季
	Season_Begin_Month        []int       //开始月份
	Season_Level_Max          uint8       //赛季档次，从0开始
	Season_Second_Level_Cnt   uint16      //每个档的小级别，1起始
	Season_Second_Level_Score uint16      //每个小级需要的积分
	Score_OneGame             []int       //单场基础积分范围
	Score_Once_Max            int         //单场最高值
	Score_Take_Off            [][]float32 //各档次失败扣除的最大积分

	// 经验
	Exp_LvUp     []uint32 //升级分布
	Exp_LvUp_Max uint32
	Exp_Once_Max uint32 //单场最高值
	Exp_Win      uint32 //胜利经验
	Exp_Fail     uint32 //失败经验
}
