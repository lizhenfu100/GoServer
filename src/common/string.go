package common

import (
	"common/std"
	"fmt"
	"strconv"
	"strings"
)

//【多字符串拼接，用bytes.Buffer.WriteString()快400-500倍】

func CheckAtoiName(s string) int {
	if len(s) <= 0 {
		fmt.Printf("CheckAtoiName: is empty")
		return 0
	}
	ret, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("CheckAtoiName can't convert to int: %s ", s)
	}
	return ret
}

// 格式：(id1|num1)(id2|num2)
func ParseStringToPair(str string) []std.IntPair {
	sFix := strings.Trim(str, "()")
	slice := strings.Split(sFix, ")(")
	items := make([]std.IntPair, len(slice))
	for i, v := range slice {
		pv := strings.Split(v, "|")
		if len(pv) < 2 {
			fmt.Printf("ParseStringToPair too short: %s", str)
			return items
		}
		items[i].ID = CheckAtoiName(pv[0])
		items[i].Cnt = CheckAtoiName(pv[1])
	}
	return items
}

// 格式：32400|43200|64800|75600
func ParseStringToArrInt(str string) []int {
	slice := strings.Split(str, "|")
	nums := make([]int, len(slice))
	for i, v := range slice {
		nums[i] = CheckAtoiName(v)
	}
	return nums
}

// -------------------------------------
// From Blizzard
var (
	crcTable            []uint32 = make([]uint32, 256)
	crcTableInitialized bool     = false
)

const crcPOLY uint32 = 0x04c11db7

func initCRCTable() {
	if crcTableInitialized {
		return
	}
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
	crcTableInitialized = true
}
func StringHash(s string) uint32 {
	initCRCTable()
	var hash, b uint32
	for _, c := range s {
		b = uint32(c)
		hash = ((hash >> 8) & 0x00FFFFFF) ^ crcTable[(hash^b)&0x000000FF]
	}
	return hash
}
