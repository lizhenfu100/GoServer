package logic

import (
	"common"
	"fmt"
	"svr_cross/api"
	"tcp"
)

//////////////////////////////////////////////////////////////////////
//! 测试msg
func Rpc_Echo(pTcpConn *tcp.TCPConn, msg *common.NetPack) {
	fmt.Println("Rpc_Echo :", msg.ReadString())

	// pTcpConn.WriteMsg(msg)
	api.SendToBattle(1, msg)
}
