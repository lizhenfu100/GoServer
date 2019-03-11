package timer

import (
	"gamelog"
	"time"
)

const (
	OneDaySecCnt = 24 * 3600
)

func IsToday(sec int64) bool { return time.Unix(sec, 0).YearDay() == time.Now().YearDay() }
func WeekInYear() int {
	_, ret := time.Now().ISOWeek()
	return ret
}
func TodayBeginSec() int64 {
	now := time.Now()
	todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return todayTime.Unix()
}
func TodayEndSec() int64 {
	return TodayBeginSec() + OneDaySecCnt
}
func TodayRunSec() int {
	now := time.Now()
	return now.Hour()*3600 + now.Minute()*60 + now.Second()
}
func TodayLeftSec() int { return OneDaySecCnt - TodayRunSec() }

// 时间戳--日期
const g_time_layout = "2006/01/02 15:04:05"

func Str2Time(date string) int64 {
	if v, err := time.ParseInLocation(g_time_layout, date, time.Local); err == nil {
		return v.Unix()
	} else {
		gamelog.Error(err.Error())
		return 0
	}
}
func Time2Str(sec int64) string { return time.Unix(sec, 0).Format(g_time_layout) }
