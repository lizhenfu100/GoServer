package std

func SetBit8(pVal *uint8, idx uint, flag bool) {
	if flag {
		(*pVal) |= (1 << idx)
	} else {
		(*pVal) &= ^(1 << idx)
	}
}
func GetBit8(val uint8, idx uint) bool { return val&(1<<idx) > 0 }

func SetBit16(pVal *uint16, idx uint, flag bool) {
	if flag {
		(*pVal) |= (1 << idx)
	} else {
		(*pVal) &= ^(1 << idx)
	}
}
func GetBit16(val uint16, idx uint) bool { return val&(1<<idx) > 0 }

func SetBit32(pVal *uint32, idx uint, flag bool) {
	if flag {
		(*pVal) |= (1 << idx)
	} else {
		(*pVal) &= ^(1 << idx)
	}
}
func GetBit32(val uint32, idx uint) bool { return val&(1<<idx) > 0 }

func SetBit64(pVal *uint64, idx uint, flag bool) {
	if flag {
		(*pVal) |= (1 << idx)
	} else {
		(*pVal) &= ^(1 << idx)
	}
}
func GetBit64(val uint64, idx uint) bool { return val&(1<<idx) > 0 }
