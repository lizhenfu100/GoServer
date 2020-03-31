package timer

import (
	"sync"
	"time"
)

type Freq struct {
	m         sync.Map
	kTotalSec byte //多少秒内
	kMaxCnt   byte //超多少次
}
type freq struct {
	sync.Mutex
	timing []int64
}

func NewFreq(maxCnt, periodSec byte) *Freq { return &Freq{kTotalSec: periodSec, kMaxCnt: maxCnt} }
func (self *Freq) Check(k interface{}) bool {
	p, _ := self.m.Load(k)
	if p == nil {
		p = &freq{}
		self.m.Store(k, p)
	}
	timenow := time.Now().Unix()
	return p.(*freq).check(timenow, self.kMaxCnt, self.kTotalSec)
}
func (self *freq) check(t int64, kMaxCnt, kTotalSec byte) bool {
	self.Lock()
	defer self.Unlock()
	if len(self.timing) == int(kMaxCnt) {
		if t-self.timing[0] < int64(kTotalSec) {
			return false
		} else {
			self.timing = append(self.timing[:0], self.timing[1:]...) //删排头，加新的
			self.timing = append(self.timing, t)
		}
	}
	self.timing = append(self.timing, t)
	return true
}
