package logic

import (
	"common/timer"
	"svr_game/logic/player"
	"time"
)

func InitTimeUpdate() {
	updateEnterNextDay()
	updateEnterNextHour()
	updatePerMinute()
}
func updateEnterNextDay() {
	delay := float32(timer.GetTodayLeftSec())
	timer.G_TimerMgr.AddTimerSec(onEnterNextDay, delay, timer.OneDay_SecCnt, -1)
}
func updateEnterNextHour() {
	now := time.Now()
	delay := float32(3600 - now.Minute()*60 - now.Second())
	timer.G_TimerMgr.AddTimerSec(onEnterNextHour, delay, 3600, -1)
}
func updatePerMinute() {
	timer.G_TimerMgr.AddTimerSec(perMinute, 0, 60, -1)
}

// ------------------------------------------------------------
// logic code
func onEnterNextDay() {
}
func onEnterNextHour() {
}
func perMinute() {
	player.NotifyOnlineNum()
}
