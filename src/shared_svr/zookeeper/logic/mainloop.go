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
	//zookeeper通告http svr消逝
	if pMeta, ok := conn.UserPtr.(*meta.Meta); ok {
		if pMeta.HttpPort > 0 {
			tcp.ForeachRegModule(func(v *tcp.TCPConn) {
				if ptr, ok := v.UserPtr.(*meta.Meta); ok {
					if pMeta.IsMyClient(ptr) || pMeta.IsMyServer(ptr) {
						v.CallRpc(enum.Rpc_http_node_quit, func(buf *common.NetPack) {
							buf.WriteString(pMeta.Module)
							buf.WriteInt(pMeta.SvrID)
						}, nil)
					}
				}
			})
		}
	}
}
