package main

import (
	"bytes"
	"common"
	"fmt"
	"gamelog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"svr_sdk/api"
	"text/template"
	"time"
)

type Leave_req struct {
	Name string
	Date string //请假时段
	Swap string //补班时段
}

// 字段含义：
//	name 姓名
//	date 请假日期
//	swap 补班日期
//
// 日期格式：4.29.17:00-19:00 | 4.29 逗号分隔表示多个日期
//
// http://192.168.1.111:7601/ask_for_leave?name=胡椒&date=4.29.17:00-19:00,4.30&swap=5.3
func Http_ask_for_leave(w http.ResponseWriter, r *http.Request) {
	//gamelog.Debug("message: %s", r.URL.String())
	r.ParseForm()

	//反射解析
	var req Leave_req
	api.Unmarshal(&req, r.Form)

	//! 创建回复
	ack := "请假失败"
	defer func() {
		w.Write([]byte(ack))
	}()

	if req.Name == "" {
		ack = "请假失败：姓名缺失"
	}
	if logInfo, ok := req.format(); ok {
		ack = logInfo.writeLog()
	} else {
		ack = "请假失败：日期格式错误 4.29.17:00-19:00"
	}
}
func (self *Leave_req) format() (ret LeaveLog, ok bool) {
	dates := strings.Split(self.Date, ",")
	swaps := strings.Split(self.Swap, ",")
	if formatDate(dates) && (self.Swap == "" || formatDate(swaps)) {
		api.CopySameField(&ret, self)
		for i := 0; i < len(dates); i++ {
			if i < len(swaps) {
				ret.List = append(ret.List, common.StrPair{dates[i], swaps[i]})
			} else {
				ret.List = append(ret.List, common.StrPair{K: dates[i]})
			}
		}
		ok = true
	}
	return
}
func formatDate(list []string) (ret bool) {
	for i := 0; i < len(list); i++ {
		//4.29.17:00-19:00 | 4.29
		if ret, _ = regexp.MatchString(`^\d{1,2}\.\d{1,2}(\.\d{1,2}:\d{1,2}-\d{1,2}:\d{1,2})?$`, list[i]); ret == false {
			gamelog.Debug("check false: %s", list[i])
			return
		} else {
			//提取日期
			reg := regexp.MustCompile(`\d{1,2}\.\d{1,2}`)
			date := strings.Split(reg.FindAllString(list[i], -1)[0], ".")
			//计算是周几
			now := time.Now()
			year := now.Year()
			month := common.CheckAtoiName(date[0])
			day := common.CheckAtoiName(date[1])
			if month == 1 && now.Month() == time.December {
				year++
			}
			date2, weekday2 := time.Date(year, time.Month(month), day, 0, 0, 0, 0, now.Location()), ""
			switch date2.Weekday() {
			case time.Monday:
				weekday2 = "周一"
			case time.Tuesday:
				weekday2 = "周二"
			case time.Wednesday:
				weekday2 = "周三"
			case time.Thursday:
				weekday2 = "周四"
			case time.Friday:
				weekday2 = "周五"
			case time.Saturday:
				weekday2 = "周六"
			case time.Sunday:
				weekday2 = "周日"
			}
			list[i] = fmt.Sprintf("%d.%d(%s)", date2.Month(), date2.Day(), weekday2)
		}
	}
	return
}

// --------------------------------------------------------------------------
// 请假信息，输出成文本
type LeaveLog struct {
	Name string
	List []common.StrPair
}

const K_Out_Template = `{
    姓名： {{.Name}}
{{range .List}}
    请假： {{.K}}  {{if ne .V ""}}补班：{{.V}}{{end}}
{{end}}
}
`

var G_OutDir = "./"

func (self *LeaveLog) writeLog() (ret string) {
	filename := time.Now().Format("2006.01") + ".afk"
	tpl, err := template.New(filename).Parse(K_Out_Template)
	if err != nil {
		panic(err.Error())
		return
	}
	var bf bytes.Buffer
	if err = tpl.Execute(&bf, self); err != nil {
		panic(err.Error())
		return
	}
	if err = os.MkdirAll(G_OutDir, 0777); err != nil {
		panic(err.Error())
		return
	}
	f, err := os.OpenFile(G_OutDir+filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
		return
	}
	defer f.Close()
	f.Write(bf.Bytes())
	return bf.String()
}
