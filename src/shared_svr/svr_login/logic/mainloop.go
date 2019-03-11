package logic

import (
	"common"
	"tcp"
)

func MainLoop() {
	InitGiftDB()
	g_bulletin.InitDB()
	AccountRegLimit()

	tcp.G_RpcQueue.Loop()
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
}
