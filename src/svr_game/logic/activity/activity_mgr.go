package activity

import (
	"common/timer"
	"time"
)

var (
	G_GlobalActivity TGlobalActivity
	g_act_timer      *timer.Timer
)

type TGlobalActivity struct {
	ActivityLst []TActivityData //! 活动列表
}
type TActivityData struct {
	ActivityID int  //! 唯一活动ID
	OpenTiems  int  //! 活动开启次数
	RunDayCnt  uint //! 本轮持续天数  【二者类型不一致，传参颠倒会编译报错】
	Status     int8 //! 状态: 0->关闭 1->开启 2->异常

	// 以下这些不必存库，初始化时读表取得
	actType   int
	beginTime int64
	endTime   int64 //! 结束时间为0的为永久运行活动
}

func (data *TActivityData) RefreshStatus(now int64) {
	data.beginTime, data.endTime = GetActivityEndTime(data.ActivityID)
	if data.endTime == 0 {
		//! 永久开启活动
		data.Status = 1
	} else if data.beginTime <= now && now <= data.endTime {
		//! 活动时间内
		data.Status = 1
	} else {
		data.Status = 0
	}
}

// ------------------------------------------------------------
// 开服，构造serv活动数据
func (self *TGlobalActivity) Init() {
	self.db_LoadGlobalActivity()
	self.CheckActivityAdd()   //! 检测表中是否有新增活动
	self.UpdateActivityTime() //! 活动开启/结束时间

	g_act_timer = timer.NewHourTimer(24)
	g_act_timer.AddTimeFunc(timer.GetTodayLeftSec(), timer.OneDay_SecCnt, -1, self)
}
func (self *TGlobalActivity) CheckActivityAdd() {
	for _, csv := range G_ActivityCsv {
		if csv.ID == 0 {
			continue
		}
		isExist := false
		for _, v := range self.ActivityLst {
			if csv.ID == v.ActivityID {
				isExist = true
				break
			}
		}
		if isExist == false {
			self.AddNewActivity(csv.ID, csv.ActivityType)
		}
	}
}
func (self *TGlobalActivity) UpdateActivityTime() {
	now := time.Now().Unix()
	for i := 0; i < len(self.ActivityLst); i++ {
		self.ActivityLst[i].RefreshStatus(now)
	}
}
func (self *TGlobalActivity) AddNewActivity(actID, actTyp int) {
	activity := TActivityData{
		ActivityID: actID,
		actType:    actTyp,
	}
	activity.RefreshStatus(time.Now().Unix())

	self.ActivityLst = append(self.ActivityLst, activity)
}
func (self *TGlobalActivity) db_LoadGlobalActivity() bool {
	return true
}

// ------------------------------------------------------------
// 活动数据刷新，在线跨天
func (self *TGlobalActivity) OnTimerRefresh(now int64) bool {
	self.EnterNextDay(now)
	return true
}
func (self *TGlobalActivity) OnTimerRunEnd(now int64) {
}
func (self *TGlobalActivity) EnterNextDay(now int64) {
	//! 【坑】range迭代的v是值拷贝，block内更改迭代数据，v的值是不变的
	//! 【坑】要是循环有先更改状态，再通过v判断的逻辑，就有问题了
	//! 【坑】涉及逻辑状态的地方，还是老实用 for i := 0... 这样的保险些~
	// for i, v := range actList {
	for i := 0; i < len(self.ActivityLst); i++ {
		v := &self.ActivityLst[i]
		if v.Status == 1 {
			if v.endTime > 0 && now >= v.endTime {
				//! 非永久 and 已过结束时间
				v.Status = 0
				v.OpenTiems += 1
				v.RunDayCnt = 0

				//! 更新下一次开始时间
				v.beginTime, v.endTime = GetActivityNextBeginTime(v.ActivityID)
			} else if v.endTime == 0 || now < v.endTime {
				//! 永久 or 没到结束时间
				v.RunDayCnt += 1
			}
		} else if v.Status == 0 {
			if now >= v.beginTime {
				//! 已经关闭的活动到达下一次开启时间
				v.Status = 1
				v.RunDayCnt += 1
			}
		}
	}
}

//! client交互时：检测重置
func (self *TActivityModule) CheckReset() {
	for _, v := range G_GlobalActivity.ActivityLst {
		pCharAct, ok := self.activityPtrs[v.ActivityID]
		if !ok || pCharAct == nil {
			continue
		}
		if v.Status == 1 && v.OpenTiems == pCharAct.OpenTiems() && v.RunDayCnt > pCharAct.RunDayCnt() {
			pCharAct.ResetDaily(v.RunDayCnt)
		} else if v.Status == 1 && v.OpenTiems != pCharAct.OpenTiems() {
			//! 活动开启轮回一次以上,需先清空上次活动数据
			pCharAct.OnEnd(v.OpenTiems, v.RunDayCnt)
			pCharAct.Init(v.ActivityID, v.OpenTiems, v.RunDayCnt)
		} else if v.Status == 0 && v.OpenTiems != pCharAct.OpenTiems() {
			pCharAct.OnEnd(v.OpenTiems, v.RunDayCnt)
		}
	}
}
