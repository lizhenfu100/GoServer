package timer

type OpFreq struct { //行为频率，如：5次/小时
	kLimitCnt int
	kTotalSec int
	timing    []int64
}

func NewOpFreq(maxCnt, periodSec int) *OpFreq {
	return &OpFreq{
		maxCnt,
		periodSec,
		nil,
	}
}
func (self *OpFreq) Check(t int64) bool {
	self.timing = append(self.timing, t)
	if len(self.timing) > self.kLimitCnt {
		if self.timing[self.kLimitCnt]-self.timing[0] > int64(self.kTotalSec) {
			self.timing = append(self.timing[:0], self.timing[1:]...)
		} else {
			return false
		}
	}
	return true
}
