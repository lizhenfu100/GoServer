package logic

import (
	"common/timer"
	"svr_game/logic/season"
	"time"
)

func InitTimeUpdate() {
	updateEnterNextDay()
	updateEnterNextHour()
	updatePerMinute()
}
func updateEnterNextDay() {
	delay := float32(timer.TodayLeftSec())
	timer.AddTimer(onEnterNextDay, delay, timer.OneDaySec, -1)
}
func updateEnterNextHour() {
	now := time.Now()
	delay := float32(3600 - now.Minute()*60 - now.Second())
	timer.AddTimer(onEnterNextHour, delay, 3600, -1)
}
func updatePerMinute() {
	timer.AddTimer(perMinute, 60, 60, -1)
}

// ------------------------------------------------------------
// logic code
func onEnterNextDay() {
	season.OnEnterNextDay()
}
func onEnterNextHour() {
}
func perMinute() {
}
