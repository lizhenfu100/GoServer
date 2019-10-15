package hash

// From Blizzard
var kCrcTable = make([]uint32, 256)

const kCrcPoly uint32 = 0x04c11db7

func init() {
	var i, j, c uint32
	for i = 0; i < uint32(len(kCrcTable)); i++ {
		c = (i << 24)
		for j = 8; j != 0; j-- {
			if (c & 0x80000000) != 0 {
				c = (c << 1) ^ kCrcPoly
			} else {
				c = (c << 1)
			}
			kCrcTable[i] = c
		}
	}
}
func StrHash(s string) uint32 {
	var hash, b uint32
	for _, c := range s {
		b = uint32(c)
		hash = ((hash >> 8) & 0x00FFFFFF) ^ kCrcTable[(hash^b)&0x000000FF]
	}
	return hash
}
