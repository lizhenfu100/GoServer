package logic

import (
	"common"
	"svr_sdk/msg"
	"tcp"
)

func MainLoop() {
	msg.InitDB()

	tcp.G_RpcQueue.Loop()
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
}
