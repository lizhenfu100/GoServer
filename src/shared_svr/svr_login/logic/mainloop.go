package logic

import (
	"common"
	"shared_svr/svr_login/gm"
	"tcp"
)

func MainLoop() {
	gm.InitGiftDB()
	gm.InitBulletin()
	AccountRegLimit()

	tcp.G_RpcQueue.Loop()
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
}
