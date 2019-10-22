package logic

import (
	"common/timer"
	"conf"
	"nets/tcp"
	"sync"
	"time"
)

func MainLoop() {
	go tcp.G_RpcQueue.Loop() //TODO:zhoumf:io线程直接转发 TCPConn.readLoop

	InitRouteGame()
	updateEnterNextDay()

	timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		timer.G_TimerMgr.Refresh(timeElapse, timeNow)
		//tcp.G_RpcQueue.Update()

		if timeElapse < conf.FPS_OtherSvr {
			time.Sleep(time.Duration(conf.FPS_OtherSvr-timeElapse) * time.Millisecond)
		}
	}
}

// ------------------------------------------------------------
// logic code
func updateEnterNextDay() {
	delay := float32(timer.TodayLeftSec() + 5*3600) //每天早上5点
	timer.G_TimerMgr.AddTimerSec(onEnterNextDay, delay, timer.OneDaySec, -1)
}
func onEnterNextDay() {
	g_route_game = sync.Map{} //清空缓存
	InitRouteGame()
}
