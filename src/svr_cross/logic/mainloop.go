package logic

import (
	"common"
	"common/timer"
	"conf"
	"netConfig/meta"
	"nets/rpc"
	"time"
)

func MainLoop() {
	for timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0; ; {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		timer.Refresh(timeElapse, timeNow)
		rpc.G_RpcQueue.Update()

		if timeElapse < conf.FPS_GameSvr {
			time.Sleep(time.Duration(conf.FPS_GameSvr-timeElapse) * time.Millisecond)
		}
	}
}
func Rpc_net_error(req, ack *common.NetPack, conn common.Conn) {
	if ptr, ok := conn.GetUser().(*meta.Meta); ok && ptr.Module == "battle" {
		delete(g_battle_player_cnt, ptr.SvrID)
	}
}
