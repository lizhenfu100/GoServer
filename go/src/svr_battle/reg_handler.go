package main

import (
	"svr_battle/logic"
	"tcp"
)

//注册http消息处理方法
func RegBattleTcpMsgHandler() {
	tcp.G_HandlerMsgMap[1] = logic.Hand_Msg_1
}
