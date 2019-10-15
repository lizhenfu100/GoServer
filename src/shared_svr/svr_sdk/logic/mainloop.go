package logic

import (
	"common"
	"common/timer"
	"conf"
	"nets/tcp"
	"shared_svr/svr_sdk/msg"
	"time"
)

func MainLoop() {
	msg.InitDB()

	updateEnterNextDay()

	timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		timer.G_TimerMgr.Refresh(timeElapse, timeNow)

		tcp.G_RpcQueue.Update()

		if timeElapse < conf.FPS_OtherSvr {
			time.Sleep(time.Duration(conf.FPS_OtherSvr-timeElapse) * time.Millisecond)
		}
	}
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
}

// ------------------------------------------------------------
// logic code
func updateEnterNextDay() {
	delay := float32(timer.TodayLeftSec())
	timer.G_TimerMgr.AddTimerSec(onEnterNextDay, delay, timer.OneDaySec, -1)
}
func onEnterNextDay() {
	ClearIpOrder()
}
