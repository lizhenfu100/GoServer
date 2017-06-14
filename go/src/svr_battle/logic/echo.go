package logic

import (
	"common"
	"fmt"
	"tcp"
)

//////////////////////////////////////////////////////////////////////
//! 测试msg
func Rpc_Echo(req, ack *common.NetPack, conn *tcp.TCPConn) {
	fmt.Println("Rpc_Echo :", req.ReadString())

	conn.WriteMsg(req)
}
