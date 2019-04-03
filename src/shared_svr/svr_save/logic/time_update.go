package logic

import (
	"common/timer"
	"shared_svr/svr_save/gm"
	"time"
)

func InitTimeUpdate() {
	updateEnterNextDay()
	updateEnterNextHour()
	updatePerMinute()
}
func updateEnterNextDay() {
	delay := float32(timer.TodayLeftSec())
	timer.G_TimerMgr.AddTimerSec(onEnterNextDay, delay, timer.OneDaySec, -1)
}
func updateEnterNextHour() {
	now := time.Now()
	delay := float32(3600 - now.Minute()*60 - now.Second())
	timer.G_TimerMgr.AddTimerSec(onEnterNextHour, delay, 3600, -1)
}
func updatePerMinute() {
	timer.G_TimerMgr.AddTimerSec(perMinute, 60, 60, -1)
}

// ------------------------------------------------------------
// logic code
func onEnterNextDay() {
	gm.G_Backup.OnEnterNextDay()
}
func onEnterNextHour() {
	gm.G_Backup.OnEnterNextHour()
}
func perMinute() {
}
