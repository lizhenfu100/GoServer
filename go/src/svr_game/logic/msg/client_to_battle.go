package msg

import (
	"common"
	"fmt"
	"svr_game/api"
)

func Rpc_Client2Battle_Echo(req, ack *common.NetPack, ptr interface{}) {
	fmt.Println(req.ReadString())

	ack.WriteString("ok")

	// 转发给Battle进程
	req.SetRpc("rpc_echo")
	api.SendToCross(req)
}
