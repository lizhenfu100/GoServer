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
package timer

import (
	"time"
)

// ------------------------------------------------------------
// 定时器：多线程写，单线程读
type timerChan struct {
	timerChan chan func()
}

func NewTimerChan(chanSize int) *timerChan { return &timerChan{make(chan func(), chanSize)} }

//ret := AfterFunc(); ret.Stop()
func (self *timerChan) AfterFunc(d time.Duration, callback func()) *time.Timer {
	//safeFun := func() {
	//	defer func() {
	//		if r := recover(); r != nil {
	//			gamelog.Error("recover Timer:%v %s", r, debug.Stack())
	//		}
	//	}()
	//	callback()
	//}
	return time.AfterFunc(d, func() {
		self.timerChan <- callback //safeFun
	})
}
func (self *timerChan) Update() {
	for {
		select {
		case cb := <-self.timerChan:
			cb()
		default:
			return
		}
	}
}
