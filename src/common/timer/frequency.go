package timer

type OpFreq struct { //行为频率，如：5次/小时
	kLimitCnt    int
	kLimitSecond int
	list         []int64
}

func NewOpFreq(maxCnt, periodSec int) *OpFreq {
	return &OpFreq{
		maxCnt,
		periodSec,
		nil,
	}
}
func (self *OpFreq) Check(t int64) bool {
	self.list = append(self.list, t)
	if len(self.list) > self.kLimitCnt {
		if self.list[self.kLimitCnt]-self.list[0] > int64(self.kLimitSecond) {
			self.list = append(self.list[:0], self.list[1:]...)
		} else {
			return false
		}
	}
	return true
}
