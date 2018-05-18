package logic

import (
	"common/timer"
	"svr_game/logic/player"
	"time"
)

func InitTimeUpdate() {
	timer.G_TimerChan = timer.NewTimerChan(40960)
	UpdateEnterNextDay()
	UpdateEnterNextHour()
}

func UpdateEnterNextDay() {
	secCnt := time.Duration(timer.GetTodayLeftSec())
	timer.G_TimerChan.AfterFunc(secCnt*time.Second, func() {
		_OnEnterNextDay()
		UpdateEnterNextDay()
	})
}
func UpdateEnterNextHour() {
	now := time.Now()
	secCnt := time.Duration(3600 - now.Minute()*60 - now.Second())
	timer.G_TimerChan.AfterFunc(secCnt*time.Second, func() {
		_OnEnterNextHour()
		UpdateEnterNextHour()
	})
}
func _OnEnterNextDay() {
	player.OnEnterNextDay_Season()
}
func _OnEnterNextHour() {

}
