package main

import (
	"svr_battle/logic"
	"tcp"
)

//注册http消息处理方法
func RegBattleTcpMsgHandler() {
	type TTcpFuncInfo struct {
		msgID uint16
		fun   func(*tcp.TCPConn, []byte)
	}
	var configSlice = []TTcpFuncInfo{
		{1, logic.Hand_Msg_1},
		{100, logic.Hand_Login},
		{101, logic.Hand_Logout},
	}

	//! register
	max := len(configSlice)
	for i := 0; i < max; i++ {
		data := &configSlice[i]
		tcp.G_HandlerMsgMap[data.msgID] = data.fun
	}
}
