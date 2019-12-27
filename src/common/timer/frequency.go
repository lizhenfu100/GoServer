package timer

import (
	"sync"
	"time"
)

type Freq struct {
	m          sync.Map
	kTotalSec  byte //多少秒内
	kMaxCnt    byte //超多少次
	kBanMinute byte //封几分钟
}
type freq struct {
	sync.Mutex
	timing []int64
}

func NewFreq(maxCnt, periodSec, banMinute byte) *Freq {
	return &Freq{
		kTotalSec:  periodSec,
		kMaxCnt:    maxCnt,
		kBanMinute: banMinute,
	}
}
func (self *Freq) Check(k interface{}) bool {
	p, _ := self.m.Load(k)
	if p == nil {
		p = &freq{}
		self.m.Store(k, p)
	}
	timenow := time.Now().Unix()
	if !p.(*freq).check(timenow, self.kMaxCnt, self.kTotalSec) {
		G_TimerMgr.AddTimerSec(func() {
			self.m.Delete(k)
		}, float32(self.kBanMinute*60), 0, 0)
		return false
	}
	return true
}
func (self *freq) check(t int64, kMaxCnt, kTotalSec byte) bool {
	self.Lock()
	self.timing = append(self.timing, t)
	if len(self.timing) > int(kMaxCnt) {
		if self.timing[kMaxCnt]-self.timing[0] > int64(kTotalSec) {
			self.timing = append(self.timing[:0], self.timing[1:]...)
		} else {
			self.Unlock()
			return false
		}
	}
	self.Unlock()
	return true
}
