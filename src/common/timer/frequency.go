package timer

import "sync"

type OpFreq struct { //行为频率，如：5次/小时
	sync.Mutex
	kLimitCnt int
	kTotalSec int
	timing    []int64
}

func NewOpFreq(maxCnt, periodSec int) *OpFreq {
	return &OpFreq{
		kLimitCnt: maxCnt,
		kTotalSec: periodSec,
	}
}
func (self *OpFreq) Check(t int64) bool {
	self.Lock()
	self.timing = append(self.timing, t)
	if len(self.timing) > self.kLimitCnt {
		if self.timing[self.kLimitCnt]-self.timing[0] > int64(self.kTotalSec) {
			self.timing = append(self.timing[:0], self.timing[1:]...)
		} else {
			self.Unlock()
			return false
		}
	}
	self.Unlock()
	return true
}
