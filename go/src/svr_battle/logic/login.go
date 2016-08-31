package logic

import (
	"tcp"
)

func Hand_Login(pTcpConn *tcp.TCPConn, buf []byte) {
	//1、从角色池取玩家
	//2、向player写入pTcpConn
}
func Hand_Logout(pTcpConn *tcp.TCPConn, buf []byte) {
	//1、置空player的pTcpConn
	//2、归还到角色池
}
