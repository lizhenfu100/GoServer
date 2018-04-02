package logic

import (
	"common"
	"tcp"
)

func MainLoop() {
	tcp.G_RpcQueue.Loop()
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
}
