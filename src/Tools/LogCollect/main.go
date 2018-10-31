package main

import (
	"common/file"
	"fmt"
	"regexp"
	"runtime/debug"
	"time"
)

var (
	//一条日志中待提取的部分；多层级匹配，依次满足所有层级的内容才会被选出
	K_Match = []string{
		`order not exists: \{\d+ \d+ \d+ \d+|order not exists: \{\d+  \d+ \d+`,
		`\{\d+ \d+ \d+ \d+|\{\d+  \d+ \d+`,
		`{\d+|\d+ \d+$`,
		`\d+`,
	}

	//输出模板
	K_Out_Template = `
{{range $_, $list := .}}
{{range $k, $v := $list}}{{if eq $k 0}}third_account:{{$v}}, {{else if eq $k 1}}order_id:{{$v}}, {{else if eq $k 2}}rmb:{{$v}}, {{end}}{{end}}
{{end}}
`

	//原始日志文件
	K_Log_Files = []string{
		"log/sdk20180404_023728.log",
	}
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%v: %s", r, debug.Stack())
			time.Sleep(time.Minute)
		}
	}()

	var result [][]string
	for _, v := range K_Log_Files {
		file.ReadLine(v, func(line string) {
			ret := matchTarget(line)
			if len(ret) > 0 && !isInResult(result, ret) {
				result = append(result, ret)
			}
		})
	}
	//fmt.Println(result)
	filename := time.Now().Format("20060102_150405") + ".log"
	file.CreateTemplate(result, "LogCollect/", filename, K_Out_Template)

	fmt.Print("Collect success...\n")
}

// 排除重复数据
func isInResult(result [][]string, val []string) bool {
	for _, v := range result {
		if len(v) == len(val) {
			isSame := true
			for i := 0; i < len(v); i++ {
				if v[i] != val[i] {
					isSame = false
					break
				}
			}
			if isSame {
				return true
			}
		}
	}
	return false
}

// 递归提取内容
func matchTarget(s string) []string { return _matchTarget(0, s) }
func _matchTarget(lv int, s string) (ret []string) {
	reg := regexp.MustCompile(K_Match[lv])
	list := reg.FindAllString(s, -1)
	lv++
	if lv < len(K_Match) {
		for _, v := range list {
			ret = append(ret, _matchTarget(lv, v)...)
		}
	} else {
		ret = append(ret, list...)
	}
	return
}
