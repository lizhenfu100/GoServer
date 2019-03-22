/***********************************************************************
* @ 定时器
* @ brief
	1、分层刷新：许多端游的做法，profiler测试没问题，逻辑简明
		· 每次主循环都调用，检查每秒时间戳，触发UpdatePerSec
		· 每秒里检查每分的时间戳，触发UpdatePerMin
		· UpdatePerHour => OnEnterNextDay
	2、timer的遍历：优先队列、小根堆、时间轮

* @ go 原生定时器    https://www.jianshu.com/p/ac94989215d6
	、time.After、time.Tick
		· 不会粗暴的增加协程，都是打包成runtimeTimer交给底层执行
	、time.AfterFunc
		· 为确保外界传入的回调不会阻塞内部定时器协程，time.AfterFunc回调都是另开协程执行的
		· time.AfterFunc -> goFunc

* @ author zhoumf
* @ date 2019-3-21
***********************************************************************/
package timer

import "common/safe"

// ------------------------------------------------------------
// 多线程写，单线程读
type SafeMgr struct {
	timeWheel
	queue safe.SafeQueue
}
type obj struct {
	ptr   *TimeNode
	delay int
}

func (self *SafeMgr) Init(cap uint32) {
	self.queue.Init(cap)
	self.timeWheel.Init()
}
func (self *SafeMgr) Refresh(timelapse int, timenow int64) {
	for {
		if v, ok, _ := self.queue.Get(); ok {
			if v := v.(obj); v.delay >= 0 {
				self.timeWheel._AddTimerNode(v.delay, v.ptr)
			} else {
				self.timeWheel.DelTimer(v.ptr)
			}
		} else {
			break
		}
	}
	self.timeWheel.Refresh(timelapse, timenow)
}
func (self *SafeMgr) DelTimer(p *TimeNode) { self.queue.Put(obj{p, -1}) }
func (self *SafeMgr) AddTimerSec(cb func(), delay, cd, total float32) *TimeNode {
	p := newNode(cb, delay, cd, total)
	self.queue.Put(obj{p, int(delay * 1000)})
	return p
}
