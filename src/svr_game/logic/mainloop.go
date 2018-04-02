package logic

import (
	"common"
	"svr_game/logic/player"
	"tcp"
	"time"
)

func MainLoop() {
	timeNow, timeOld, time_elapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		time_elapse = int(timeNow - timeOld)

		player.G_ServiceMgr.RunAllService(time_elapse, timeNow)

		tcp.G_RpcQueue.Update()

		time.Sleep(50 * time.Millisecond)
	}
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
}
