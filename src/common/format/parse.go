package format

import (
	"common/std"
	"regexp"
	"strconv"
	"strings"
)

// 格式：(id1|num1)(id2|num2)
func StrToPair(str string) []std.IntPair {
	sFix := strings.Trim(str, "()")
	slice := strings.Split(sFix, ")(")
	ret := make([]std.IntPair, len(slice))
	for i, v := range slice {
		if sub := strings.Split(v, "|"); len(sub) >= 2 {
			ret[i].ID, _ = strconv.Atoi(sub[0])
			ret[i].Cnt, _ = strconv.Atoi(sub[1])
		} else {
			println("StrToPair too short: ", v)
		}
	}
	return ret
}

// 格式：32400|43200|64800|75600
func StrToInts(str string) []int {
	slice := strings.Split(str, "|")
	ret := make([]int, len(slice))
	for i, v := range slice {
		ret[i], _ = strconv.Atoi(v)
	}
	return ret
}

// 连续空白字符，合并成一个空格
func MergeNearSpace(str string) string {
	reg := regexp.MustCompile(`\s+`)
	return reg.ReplaceAllString(str, " ")
}
