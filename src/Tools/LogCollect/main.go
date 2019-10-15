package main

import (
	"common/file"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

//var (
//	//一条日志中待提取的部分；多层级匹配，依次满足所有层级的内容才会被选出
//	K_Match = []string{
//		`order not exists: \{\d+ \d+ \d+ \d+|order not exists: \{\d+  \d+ \d+`,
//		`\{\d+ \d+ \d+ \d+|\{\d+  \d+ \d+`,
//		`{\d+|\d+ \d+$`,
//		`\d+`,
//	}
//	K_Out_Template = `
//{{range $_, $list := .}}
//{{range $k, $v := $list}}{{if eq $k 0}}third_account:{{$v}}, {{else if eq $k 1}}order_id:{{$v}}, {{else if eq $k 2}}rmb:{{$v}}, {{end}}{{end}}
//{{end}}
//`
//)
var g_conf struct {
	Match    []string //一条日志中待提取的部分；多层级匹配，依次满足所有层级的内容才会被选出
	Template string   //输出模板
	Dedup    bool     //去重否
}

func main() {
	_file, _day := "", 0
	flag.StringVar(&_file, "f", "", "file list")
	flag.IntVar(&_day, "d", 3, "number of days")
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%v: %s", r, debug.Stack())
			time.Sleep(time.Minute)
		}
	}()
	file.LoadCsv("csv/log_match.csv", &g_conf)

	var files []string
	if _file != "" {
		files = strings.Split(_file, " ")
	} else {
		names, _ := file.WalkDir("log/", ".log") //扫描近几天的所有日志
		for _, v := range names {
			f, _ := os.Open(v)
			fi, _ := f.Stat()
			if time.Now().Sub(fi.ModTime()) <= time.Duration(_day)*time.Hour*24 {
				files = append(files, v)
			}
		}
	}
	var result [][]string
	for _, v := range files {
		file.ReadLine(v, func(line string) {
			ret := matchTarget(line)
			if len(ret) > 0 && !isInResult(result, ret) {
				result = append(result, ret)
			}
		})
	}
	//fmt.Println(result)
	file.CreateTemplate(result, "./", "log.out", g_conf.Template)
	fmt.Println("Collect success...")
}

// 排除重复数据
func isInResult(result [][]string, val []string) bool {
	if g_conf.Dedup {
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
	}
	return false
}

// 递归提取内容
func matchTarget(s string) []string { return _matchTarget(0, s) }
func _matchTarget(lv int, s string) (ret []string) {
	reg := regexp.MustCompile(g_conf.Match[lv])
	list := reg.FindAllString(s, -1)
	lv++
	if lv < len(g_conf.Match) {
		for _, v := range list {
			ret = append(ret, _matchTarget(lv, v)...)
		}
	} else {
		ret = append(ret, list...)
	}
	return
}
