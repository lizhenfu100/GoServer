package logic

import (
	"common"
	"tcp"
)

func Rpc_Login(pTcpConn *tcp.TCPConn, msg *common.NetPack) {
	//1、从角色池取玩家
	//2、向player写入pTcpConn
}
func Rpc_Logout(pTcpConn *tcp.TCPConn, msg *common.NetPack) {
	//1、置空player的pTcpConn
	//2、归还到角色池
}
