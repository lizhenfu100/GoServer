package logic

import (
	"time"

	"svr_game/logic/player"
)

func MainLoop() {
	// init func list
	player.InitSvrMailLst()

	timeOld, timeNow, time_elasped := 0, time.Now().Nanosecond()/int(time.Millisecond), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().Nanosecond() / int(time.Millisecond)
		time_elasped = timeNow - timeOld

		player.G_Auto_Write_DB.RunSevice(time_elasped)

		time.Sleep(100 * time.Millisecond)
	}
}
