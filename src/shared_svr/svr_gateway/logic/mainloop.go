package logic

import (
	"common"
	"common/timer"
	"conf"
	"generate_out/rpc/enum"
	"netConfig/meta"
	"nets/tcp"
	"time"
)

func MainLoop() {
	go tcp.G_RpcQueue.Loop()

	timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		timer.G_TimerMgr.Refresh(timeElapse, timeNow)

		if timeElapse < conf.FPS_OtherSvr {
			time.Sleep(time.Duration(conf.FPS_OtherSvr-timeElapse) * time.Millisecond)
		}
	}
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
	if accountId, ok := conn.UserPtr.(uint32); ok { //玩家离线
		//通知游戏服
		if p := GetGameConn(accountId); p != nil {
			p.CallRpc(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
				buf.WriteUInt16(enum.Rpc_game_logout)
				buf.WriteUInt32(accountId)
			}, nil)
		}
		//清空缓存
		DelClientConn(accountId)
		DelGameConn(accountId)
	} else if ptr, ok := conn.UserPtr.(*meta.Meta); ok && ptr.Module == "game" { //游戏服断开
	}
}
