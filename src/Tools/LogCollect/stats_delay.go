package main

import (
	"common/file"
	"encoding/json"
	"fmt"
	"sort"
)

//var K_Match = []string{
//	`\{\S*\}`,
//	`[^\{\}("location":)("isp":)("delay":)]`,
//}

// ------------------------------------------------------------
type StringSlice [][]string

func (p StringSlice) Len() int      { return len(p) }
func (p StringSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p StringSlice) Less(i, j int) bool {
	if p[i][0] == p[j][0] {
		return p[i][1] < p[j][1]
	} else {
		return p[i][0] < p[j][0]
	}
}

type Info struct {
	Area  string
	Delay int
}

func DoStats(result StringSlice) { //统计外网延时log
	sort.Sort(result)
	var infos []Info
	stats := make([]int, 100)
	file.ParseCsv(result, &infos)
	ret := map[string]int{}
	cur, delay, n := "", 0, 0
	for _, v := range infos {
		//fmt.Println(v)
		if isSame(v.Area, cur) {
			n++
			cur = v.Area
			delay += v.Delay
			if v.Delay/1000 >= len(stats) {
				fmt.Println(v)
			} else {
				stats[v.Delay/1000]++
			}
		} else {
			if cur != "" {
				ret[cur] = delay / n

				for i, v := range stats {
					stats[i] = v * 100 / n
				}
				fmt.Println(cur, stats)
			}
			stats = make([]int, 100)
			cur, delay, n = v.Area, v.Delay, 1
		}
	}
	if cur != "" {
		ret[cur] = delay / n
	}
	b, _ := json.MarshalIndent(ret, "", "     ")
	fmt.Println(string(b))
}
func isSame(a, b string) bool {
	const kLen = len("四川")
	if len(a) >= kLen && len(b) >= kLen {
		return a[:kLen] == b[:kLen]
	}
	return a == b
}
