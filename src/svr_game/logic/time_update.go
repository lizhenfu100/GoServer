package logic

import (
	"common/timer"
	"svr_game/logic/player"
	"time"
)

func InitTimeUpdate(t *timer.TimerChan) {
	UpdateEnterNextDay(t)
	UpdateEnterNextHour(t)
}

func UpdateEnterNextDay(t *timer.TimerChan) {
	secCnt := time.Duration(timer.GetTodayLeftSec())
	t.AfterFunc(secCnt*time.Second, func() {
		_OnEnterNextDay()
		UpdateEnterNextDay(t)
	})
}
func UpdateEnterNextHour(t *timer.TimerChan) {
	now := time.Now()
	secCnt := time.Duration(3600 - now.Minute()*60 - now.Second())
	t.AfterFunc(secCnt*time.Second, func() {
		_OnEnterNextHour()
		UpdateEnterNextHour(t)
	})
}
func _OnEnterNextDay() {
	player.OnEnterNextDay_Season()
}
func _OnEnterNextHour() {

}
