package tcp

import (
	"fmt"
)

//Notice：http的消息处理，是另开goroutine调用的，所以函数中可阻塞；tcp就不行了
func RegTcpMsgHandler() {
	HandleFunc(1, Hand_Msg_1)
}

//////////////////////////////////////////////////////////////////////
//! 测试msg
//////////////////////////////////////////////////////////////////////
func Hand_Msg_1(pTcpConn *TCPConn, buf []byte) {
	fmt.Println("Hand_Msg_1:", buf)
}
