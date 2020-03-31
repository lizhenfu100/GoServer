package logic

import (
	"common/timer"
	"common/tool/usage"
	"conf"
	"shared_svr/svr_center/account"
	"time"
)

func MainLoop() {
	account.Init()
	timer.AddTimer(usage.Check, 60, 600, -1)

	for timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0; ; {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		timer.Refresh(timeElapse, timeNow)

		if timeElapse < conf.FPS_OtherSvr {
			time.Sleep(time.Duration(conf.FPS_OtherSvr-timeElapse) * time.Millisecond)
		}
	}
}
