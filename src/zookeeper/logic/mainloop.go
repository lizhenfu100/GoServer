package logic

import (
	"common"
	"common/net/meta"
	"generate_out/rpc/enum"
	"tcp"
	"time"
)

func MainLoop() {
	//timeNow, timeOld, time_elapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		//timeOld = timeNow
		//timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		//time_elapse = int(timeNow - timeOld)

		tcp.G_RpcQueue.Update()

		time.Sleep(10 * time.Millisecond)
	}
}
func Rpc_report_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
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
