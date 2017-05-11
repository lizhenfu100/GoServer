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
	"sync"
	"time"
)

func IsToday(day int) bool { return time.Now().Day() == day }
func WeekInYear() int {
	_, ret := time.Now().ISOWeek()
	return ret
}
func GetTodayBeginSec() int64 {
	now := time.Now()
	todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return todayTime.Unix()
}
func GetTodayEndSec() int64 {
	return GetTodayBeginSec() + OneDay_SecCnt
}
func GetTodayRunSec() int64 {
	now := time.Now()
	return int64(now.Hour()*3600 + now.Minute()*60 + now.Second())
}
func GetTodayLeftSec() int64 {
	return OneDay_SecCnt - GetTodayRunSec()
}

const (
	OneDay_SecCnt = 24 * 3600
	INT_MAX       = 0x7FFFFFFF
)

type TimeHandler interface {
	OnTimerRefresh(now int64) bool //失败即停止循环
	OnTimerRunEnd(now int64)
}
type TimerFunc struct {
	cdSec    int   // 间隔多久
	runSec   int   // 总共跑多久
	dealTime int64 // 要处理的时刻点
	handler  TimeHandler
}
type Timer struct {
	sync.Mutex
	funcLst []TimerFunc
	addLst  []TimerFunc
	delLst  []TimerFunc
}

//! 计时器
func (self *Timer) _OnTimer(interval time.Duration) {
	timer := time.NewTimer(interval)
	for {
		select {
		case <-timer.C:
			now := time.Now().Unix()
			self._OnTimerFunc(now)
			timer.Reset(time.Second)
		}
	}
}
func (self *Timer) _OnTimerFunc(now int64) {
	// 先处理要增、删的项
	self.Lock()
	{
		self.funcLst = append(self.funcLst, self.addLst...)
		for i := 0; i < len(self.delLst); i++ {
			for j := len(self.funcLst) - 1; j >= 0; j-- { // 倒过来找，快一点
				if self.funcLst[j].handler == self.delLst[i].handler {
					self.funcLst = append(self.funcLst[:j], self.funcLst[j+1:]...)
				}
			}
		}
	}
	self.Unlock()

	//TODO：funcLst遍历优化
	isDelete := false
	for i := 0; i < len(self.funcLst); i++ { // 这里得用len()，每次迭代都算长度，因为可能删除
		data := &self.funcLst[i]
		if now >= data.dealTime {
			if data.handler.OnTimerRefresh(now) {

				isDelete = false
				data.dealTime += int64(data.cdSec)

				if data.runSec != INT_MAX {
					if data.runSec -= data.cdSec; data.runSec < 0 { //! 注意差一Bug，==0还要跑一次
						isDelete = true
					}
				}
			} else {
				isDelete = true
			}
			if isDelete {
				data.handler.OnTimerRunEnd(now)
				self.funcLst = append(self.funcLst[:i], self.funcLst[i+1:]...)
				i--
				//! 删除该timer：上面的data仍在引用，在c++中若后续data指针被使用，就野了
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////
// API
//////////////////////////////////////////////////////////////////////
func NewHourTimer(interval time.Duration) *Timer {
	return _Newimer(interval * time.Hour)
}
func NewMinuteTimer(interval time.Duration) *Timer {
	return _Newimer(interval * time.Minute)
}
func NewSecondTimer(interval time.Duration) *Timer {
	return _Newimer(interval * time.Second)
}
func _Newimer(interval time.Duration) *Timer {
	p := new(Timer)
	go p._OnTimer(interval)
	return p
}

//  延时多久后开始执行
//! 用于简单逻辑调用：可重复添加
func (self *Timer) AddTimeFunc(delaySec int64, cdSec, runSec int, handler TimeHandler) {
	if runSec < 0 {
		runSec = INT_MAX
	}
	nextDealTime := time.Now().Unix() + delaySec

	self.Lock()
	defer self.Unlock()
	self.addLst = append(self.addLst, TimerFunc{cdSec, runSec, nextDealTime, handler})
}
func (self *Timer) DelTimeFunc(handler TimeHandler) {
	self.Lock()
	defer self.Unlock()
	self.delLst = append(self.delLst, TimerFunc{handler: handler})
}

func (self *Timer) AddTimeFunc_S(delaySec int64, cdSec, runSec int, handler TimeHandler) {
	if runSec < 0 {
		runSec = INT_MAX
	}
	nextDealTime := time.Now().Unix() + delaySec

	self.Lock()
	defer self.Unlock()

	for i := 0; i < len(self.funcLst); i++ {
		if self.funcLst[i].handler == handler { //已有
			return
		}
	}
	for i := 0; i < len(self.addLst); i++ {
		if self.addLst[i].handler == handler { //已有
			return
		}
	}
	self.addLst = append(self.addLst, TimerFunc{cdSec, INT_MAX, nextDealTime, handler})
}
