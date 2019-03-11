package logic

import (
	"common"
	"shared_svr/svr_center/account"
	"tcp"
)

func MainLoop() {
	account.InitDB()

	tcp.G_RpcQueue.Loop()
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
}
