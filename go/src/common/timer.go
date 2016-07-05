/***********************************************************************
* @ 定时器
* @ brief
	1、分层Update：
		· (许多端游的做法，profiler测试没问题，逻辑简明)
		· 每次主循环都调用，检查每秒时间戳，触发UpdatePerSec
		· 每秒里检查每分的时间戳，触发UpdatePerMin
		· UpdatePerHour => OnEnterNextDay
	2、timer的遍历可用：优先队列、小根堆、时间轮
* @ author zhoumf
* @ date 2016-6-29
***********************************************************************/
package common

import (
	"time"
)

func GetTodayStartSec() int64 {
	now := time.Now()
	todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return todayTime.Unix()
}

const (
	OneDay_SecCnt = 24 * 60 * 60
	INT_MAX       = 999999999
)

type DealFunc func(now int64) bool // 处理失败即停止循环
type TimerFunc struct {
	FuncID    uint
	CDSec     int // 间隔多久
	RunSec    int // 总共跑多久
	ResetTime int64
	deal      DealFunc
}
type Timer struct {
	FuncLst []TimerFunc
}

//! 计时器
func (self *Timer) OnTimer() {
	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-timer.C:
			now := time.Now().Unix()
			self.OnTimerFunc(now)
			timer.Reset(time.Second)
		}
	}
}
func (self *Timer) OnTimerFunc(now int64) {
	isDelete := false
	for i := 0; i < len(self.FuncLst); i++ { //每秒遍历【优化】
		data := &self.FuncLst[i]
		if now >= data.ResetTime {
			if data.deal(now) {
				isDelete = false
				data.ResetTime += int64(data.CDSec)
				if data.RunSec != INT_MAX {
					data.RunSec -= data.CDSec
					if data.RunSec < 0 { //! 注意差一Bug，==0还要跑一次
						isDelete = true
					}
				}
			} else {
				isDelete = true
			}

			if isDelete {
				self.FuncLst = append(self.FuncLst[:i], self.FuncLst[i+1:]...)
				i--
				//! 删除该timer：上面的data仍在引用，在c++中若后续data指针被使用，就野了
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////
// API
//////////////////////////////////////////////////////////////////////
var G_Timer Timer

func (self *Timer) Init() {

	resetTime := GetTodayStartSec() + OneDay_SecCnt
	self.AddTimeFuncByID(Timer_FuncID_Activity, resetTime, OneDay_SecCnt, nil)

	go self.OnTimer()
}

const (
	Timer_FuncID_Arena    = 1 //! 竞技场
	Timer_FuncID_Activity = 6 //! 活动刷新
)

//! 用于系统刷新调用：会替换同ID的处理函数
func (self *Timer) AddTimeFuncByID(funcID uint, resetTime int64, cdSec int, deal func(int64) bool) {
	for i := 0; i < len(self.FuncLst); i++ {
		if self.FuncLst[i].FuncID == funcID { //已有
			self.FuncLst[i].deal = deal
			return
		}
	}
	self.FuncLst = append(self.FuncLst, TimerFunc{funcID, cdSec, INT_MAX, resetTime, deal})
}

//  延时多久后开始执行
//! 用于简单逻辑调用：可重复添加，自己保证不会一轮跑多次
func (self *Timer) AddTimeFunc(delaySec int64, cdSec, runSec int, deal func(int64) bool) {
	resetTime := time.Now().Unix() + delaySec
	self.FuncLst = append(self.FuncLst, TimerFunc{0, cdSec, runSec, resetTime, deal})
}
