package logic

import (
	"fmt"
	"svr_battle/api"
	"tcp"

	"unsafe"
)

//////////////////////////////////////////////////////////////////////
//! 测试msg
func Hand_Msg_1(pTcpConn *tcp.TCPConn, buf []byte) {
	fmt.Println("Hand_Msg_1 :", *(*string)(unsafe.Pointer(&buf)))

	// pTcpConn.WriteMsg(1, buf)
	api.SendToGame(1, buf)
}
