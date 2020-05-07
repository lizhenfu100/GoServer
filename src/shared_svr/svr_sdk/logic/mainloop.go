package logic

import (
	"common/timer"
	"common/tool/usage"
	"conf"
	"nets/tcp"
	"shared_svr/svr_sdk/msg"
	"time"
)

func MainLoop() {
	updateEnterNextDay()
	timer.AddTimer(usage.Check, 60, 600, -1)
	for timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0; ; {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		timer.Refresh(timeElapse, timeNow)
		tcp.G_RpcQueue.Update()

		if timeElapse < conf.FPS_OtherSvr {
			time.Sleep(time.Duration(conf.FPS_OtherSvr-timeElapse) * time.Millisecond)
		}
	}
}

// ------------------------------------------------------------
// logic code
func updateEnterNextDay() {
	delay := float32(timer.TodayLeftSec())
	timer.AddTimer(onEnterNextDay, delay, timer.OneDaySec, -1)
}
func onEnterNextDay() {
	msg.ClearOldOrder()
	ClearMacOrder()
}
