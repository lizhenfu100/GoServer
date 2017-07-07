package logic

import (
	"time"

	"svr_game/logic/player"
)

func MainLoop() {
	// init func list
	player.InitSvrMailLst()

	timeNow, timeOld, time_elapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		time_elapse = int(timeNow - timeOld)

		player.G_Service_Write_DB.RunSevice(time_elapse)
		player.G_Service_Check_AFK.RunSevice(timeNow)

		time.Sleep(1000 * time.Millisecond)
	}
}
