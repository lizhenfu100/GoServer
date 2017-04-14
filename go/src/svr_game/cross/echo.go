package cross

import (
	"common"
	"fmt"
	"tcp"
)

//////////////////////////////////////////////////////////////////////
//! 测试msg
func Rpc_Echo(pTcpConn *tcp.TCPConn, msg *common.NetPack) {
	fmt.Println("Rpc_Echo :", msg.ReadString())

	// pTcpConn.WriteMsg(msg)
}
