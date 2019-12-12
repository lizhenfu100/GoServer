package http

import (
	"common"
	"generate_out/err"
	"shared_svr/svr_gateway/logic"
	"time"
)

func Rpc_gateway_login(req, ack *common.NetPack) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()

	if logic.CheckToken(accountId, token) {
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Token_verify_err)
	}
}
func Rpc_gateway_login_token(req, ack *common.NetPack) {
	logic.Rpc_gateway_login_token(req)
}

func Rpc_gateway_relay_module(req, ack *common.NetPack) {
	logic.Rpc_gateway_relay_module(req, func(backBuf *common.NetPack) {
		ack.WriteBuf(backBuf.LeftBuf())
	})
}
func Rpc_gateway_relay_modules(req, ack *common.NetPack) {
	logic.Rpc_gateway_relay_modules(req, func(backBuf *common.NetPack) {
		ack.WriteBuf(backBuf.LeftBuf())
	})
}
func Rpc_gateway_relay_player_msg(req, ack *common.NetPack) {
	logic.Rpc_gateway_relay_player_msg(req, func(backBuf *common.NetPack) {
		ack.WriteBuf(backBuf.LeftBuf())
	})
}
func Rpc_timestamp(req, ack *common.NetPack) {
	ack.WriteInt64(time.Now().Unix())
}
