package common

type TBitFlag struct {
	BitFlag int8
}

func (self *TBitFlag) GetBitFlag(idx uint) bool {
	return self.BitFlag&(1<<idx) > 0
}
func (self *TBitFlag) SetBitFlag(idx uint, flag bool) {
	if flag {
		self.BitFlag |= (1 << idx)
	} else {
		self.BitFlag &= ^(1 << idx)
	}
}
