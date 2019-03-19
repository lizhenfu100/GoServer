package hash

// From Blizzard
var (
	crcTable = make([]uint32, 256)
	crcInit  = false
)

const crcPOLY uint32 = 0x04c11db7

func initCRCTable() {
	if !crcInit {
		var i, j, c uint32
		for i = 0; i < 256; i++ {
			c = (i << 24)
			for j = 8; j != 0; j-- {
				if (c & 0x80000000) != 0 {
					c = (c << 1) ^ crcPOLY
				} else {
					c = (c << 1)
				}
				crcTable[i] = c
			}
		}
		crcInit = true
	}
}
func StrHash(s string) uint32 {
	initCRCTable()
	var hash, b uint32
	for _, c := range s {
		b = uint32(c)
		hash = ((hash >> 8) & 0x00FFFFFF) ^ crcTable[(hash^b)&0x000000FF]
	}
	return hash
}
