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
	INT_MAX       = 0xFFFFFFFF
)

type TimeHandler interface {
	OnTimerRefresh(now int64) bool //失败即停止循环
	OnTimerRunEnd(now int64)
}
type TimerFunc struct {
	CDSec    int   // 间隔多久
	RunSec   int   // 总共跑多久
	DealTime int64 // 要处理的时刻点
	handler  TimeHandler
}
type Timer struct {
	mutex   sync.Mutex
	isLock  bool
	funcLst []TimerFunc
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
	var list []TimerFunc
	if self.isLock {
		// copy on write
		self.mutex.Lock()
		list = append(list, self.funcLst...)
		self.mutex.Unlock()
	} else {
		list = self.funcLst
	}

	isDelete, max := false, len(list)
	for i := 0; i < max; i++ { //TODO：每秒遍历【优化】
		data := &list[i]
		if now >= data.DealTime {
			if data.handler.OnTimerRefresh(now) {

				isDelete = false
				data.DealTime += int64(data.CDSec)

				if data.RunSec != INT_MAX {
					if data.RunSec -= data.CDSec; data.RunSec < 0 { //! 注意差一Bug，==0还要跑一次
						isDelete = true
					}
				}
			} else {
				isDelete = true
			}
			if isDelete {
				data.handler.OnTimerRunEnd(now)
				list = append(list[:i], list[i+1:]...)
				i--
				//! 删除该timer：上面的data仍在引用，在c++中若后续data指针被使用，就野了
			}
		}
	}
}
func (self *Timer) _Lock() {
	self.mutex.Lock()
	self.isLock = true
}
func (self *Timer) _Unlock() {
	self.mutex.Unlock()
	self.isLock = false
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

	self._Lock()
	defer self._Unlock()
	self.funcLst = append(self.funcLst, TimerFunc{cdSec, runSec, nextDealTime, handler})
}
func (self *Timer) DelTimeFunc(handler TimeHandler) {
	self._Lock()
	defer self._Unlock()

	for i := 0; i < len(self.funcLst); i++ {
		if self.funcLst[i].handler == handler {
			self.funcLst = append(self.funcLst[:i], self.funcLst[i+1:]...)
			return
		}
	}
}

func (self *Timer) AddTimeFunc_S(delaySec int64, cdSec, runSec int, handler TimeHandler) {
	if runSec < 0 {
		runSec = INT_MAX
	}
	nextDealTime := time.Now().Unix() + delaySec

	self._Lock()
	defer self._Unlock()

	for i := 0; i < len(self.funcLst); i++ {
		if self.funcLst[i].handler == handler { //已有
			return
		}
	}
	self.funcLst = append(self.funcLst, TimerFunc{cdSec, INT_MAX, nextDealTime, handler})
}
