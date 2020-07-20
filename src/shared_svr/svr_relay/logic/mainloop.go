package logic

import (
	"common/timer"
	"conf"
	"shared_svr/svr_relay/logic/lockstep"
	"shared_svr/svr_relay/player"
	"time"
)

func MainLoop() {
	for timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0; ; {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		player.G_Service.RunSevice(timeElapse, timeNow)
		timer.Refresh(timeElapse, timeNow)
		lockstep.G_test.Broadcast()

		if timeElapse < conf.FPS_GameSvr {
			time.Sleep(time.Duration(conf.FPS_GameSvr-timeElapse) * time.Millisecond)
		}
	}
}
