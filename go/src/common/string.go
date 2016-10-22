package common

import (
	"fmt"
	"strconv"
	"strings"
)

func CheckAtoiName(s string) int {
	if len(s) <= 0 {
		fmt.Printf("field: %s is empty", s)
		return 0
	}
	ret, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("field: %s text can't convert to int", s)
	}
	return ret
}

// 格式：(id1|num1)(id2|num2)
func ParseStringToPair(str string) []IntPair {
	sFix := strings.Trim(str, "()")
	slice := strings.Split(sFix, ")(")
	items := make([]IntPair, len(slice))
	for i, v := range slice {
		pv := strings.Split(v, "|")
		if len(pv) != 2 {
			fmt.Printf("ParseStringToPair : %s", str)
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
