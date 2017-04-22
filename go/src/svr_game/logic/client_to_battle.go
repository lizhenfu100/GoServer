package logic

import (
	"common"
	"fmt"
	"svr_game/api"
)

//! 消息处理函数
//
func Handle_Client2Battle_Echo(req, ack *common.NetPack) {
	fmt.Println(req.ReadString())

	ack.WriteString("ok")

	// 转发给Battle进程
	// api.SendToBattle(1, req)
	api.SendToCross(req)
}
