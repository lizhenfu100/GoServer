package logic

import (
	"time"

	"svr_game/logic/player"
)

func MainLoop() {
	// init func list
	player.InitSvrMailLst()

	timeOld, timeNow := 0, time.Now().Nanosecond()/int(time.Millisecond)
	for {
		timeOld = timeNow
		timeNow = time.Now().Nanosecond() / int(time.Millisecond)

		player.G_Auto_Write_DB.RunSevice(timeNow - timeOld)

		time.Sleep(100 * time.Millisecond)
	}
}
