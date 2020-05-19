package logic

import (
	"common/timer"
	"common/tool/usage"
	"conf"
	"nets/rpc"
	"svr_game/player"
	"time"
)

func MainLoop() {
	player.InitDB()
	InitTimeUpdate()
	timer.AddTimer(usage.Check, 60, 600, -1)

	for timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0; ; {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		player.G_ServiceMgr.RunAllService(timeElapse, timeNow)
		timer.Refresh(timeElapse, timeNow)
		rpc.G_RpcQueue.Update()

		if timeElapse < conf.FPS_GameSvr {
			time.Sleep(time.Duration(conf.FPS_GameSvr-timeElapse) * time.Millisecond)
		}
	}
}
