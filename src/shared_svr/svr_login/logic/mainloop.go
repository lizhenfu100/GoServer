package logic

import (
	"common/timer"
	"conf"
	"nets/http"
	"nets/rpc"
	"time"
)

func MainLoop() {
	http.SetIntercept(AccountLimit)

	for timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0; ; {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		timer.Refresh(timeElapse, timeNow)
		rpc.G_RpcQueue.Update()

		if timeElapse < conf.FPS_OtherSvr {
			time.Sleep(time.Duration(conf.FPS_OtherSvr-timeElapse) * time.Millisecond)
		}
	}
}
