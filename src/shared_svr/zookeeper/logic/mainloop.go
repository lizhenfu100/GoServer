package logic

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig/meta"
	"tcp"
)

func MainLoop() {
	tcp.G_RpcQueue.Loop()
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
