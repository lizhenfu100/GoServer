package logic

import (
	"common"
	"fmt"
	"svr_game/api"
)

func Rpc_Client2Battle_Echo(req, ack *common.ByteBuffer) interface{} {
	fmt.Println(req.ReadString())

	ack.WriteString("ok")

	// 转发给Battle进程
	msg := common.NewNetPackCap(req.Size() + common.PACK_HEADER_SIZE)
	msg.SetRpc("rpc_echo")
	msg.WriteBuf(req.DataPtr)
	// api.SendToBattle(1, msg)
	api.SendToCross(msg)

	return nil
}
