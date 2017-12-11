package logic

import (
	"fmt"
	"svr_game/logic/player"
	"time"
)

var (
	G_TriggerCsv  map[int]*TriggerCsv
	g_trigger_fun = []func(*player.TPlayer, int, int) bool{
		nil,          //0
		_during_time, //1
	}
)

type TriggerCsv struct {
	ID     int
	Type   int
	Value1 int //1223145632：12月23号14:56:32
	Value2 int
}

func Check(player *player.TPlayer, ids ...int) bool {
	for _, id := range ids {
		csv := G_TriggerCsv[id]
		if csv == nil || !g_trigger_fun[csv.Type](player, csv.Value1, csv.Value2) {
			return false
		}
	}
	return true
}

func _during_time(player *player.TPlayer, val1, val2 int) bool {
	now := time.Now()
	timeNow, yearNow := now.Unix(), now.Year()
	var time1, time2 int64
	if tt, err := time.Parse("20060102150405", fmt.Sprintf("%d%d", yearNow, val1)); err == nil {
		time1 = tt.Unix()
	}
	if tt, err := time.Parse("20060102150405", fmt.Sprintf("%d%d", yearNow, val2)); err == nil {
		time2 = tt.Unix()
	}
	return timeNow > time1 && timeNow < time2
}
