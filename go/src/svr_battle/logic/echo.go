package logic

import (
	"fmt"
	"netConfig"
	"tcp"

	"unsafe"
)

var (
	g_cache_game_conn *tcp.TCPConn
)

func SendToGame(msgID uint16, msgdata []byte) {
	if g_cache_game_conn == nil {
		g_cache_game_conn = netConfig.GetTcpConn("game", 0)
	}
	g_cache_game_conn.WriteMsg(msgID, msgdata)
}

//////////////////////////////////////////////////////////////////////
//! 测试msg
func Hand_Msg_1(pTcpConn *tcp.TCPConn, buf []byte) {
	fmt.Println("Hand_Msg_1 :", *(*string)(unsafe.Pointer(&buf)))

	// pTcpConn.WriteMsg(1, buf)
	SendToGame(1, buf)
}
