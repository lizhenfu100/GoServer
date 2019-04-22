package activity

//! 活动通用接口
type FunActivity interface {
	Init(actID int, openTiems int, runDayCnt uint) // Init带有actID可将一份活动数据重置为另一活动，如同种活动Typ不同奖励
	OnEnd(openTiems int, runDayCnt uint)
	ResetDaily(runDayCnt uint)
	OpenTiems() int
	RunDayCnt() uint
}
type TActivityModule struct { // 玩家的活动数据
	// ……
	// ……

	activityPtrs map[int]FunActivity //! 各类活动Init时注册到这里
}
