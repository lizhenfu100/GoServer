package http

import (
	"common"
	"generate_out/err"
	"shared_svr/svr_gateway/logic"
)

func Rpc_check_identity(req, ack *common.NetPack) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()
	if logic.CheckToken(accountId, token) {
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Token_verify_err)
	}
}
func Rpc_set_identity(req, ack *common.NetPack) { logic.Rpc_set_identity(req) }
func Rpc_gateway_relay(req, ack *common.NetPack) {
	logic.Rpc_gateway_relay(req, func(backBuf *common.NetPack) {
		ack.ResetHead(backBuf)
		ack.WriteBuf(backBuf.LeftBuf())
	})
}
