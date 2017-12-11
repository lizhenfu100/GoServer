package logic

import (
	"time"

	"svr_game/logic/player"
)

func MainLoop() {
	timeNow, timeOld, time_elapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		time_elapse = int(timeNow - timeOld)

		player.G_ServiceMgr.RunAllService(time_elapse, timeNow)

		time.Sleep(1000 * time.Millisecond)
	}
}
