package logic

import (
	"common"
	"netConfig"
	"netConfig/meta"
	"tcp"
)

func Rpc_gateway_relay_module(req, ack *common.NetPack, conn *tcp.TCPConn) {
	module := req.ReadString()
	svrId := req.ReadInt()
	rpcId := req.ReadUInt16()
	oldReqKey := req.GetReqKey()

	if p := netConfig.GetTcpConn(module, svrId); p != nil {
		p.CallRpc(rpcId, func(buf *common.NetPack) {
			buf.WriteBuf(req.LeftBuf())
		}, func(backBuf *common.NetPack) {
			//异步回调，不能直接用ack
			backBuf.SetReqKey(oldReqKey)
			conn.WriteMsg(backBuf)
		})
	}
}
func Rpc_gateway_relay_modules(req, ack *common.NetPack, conn *tcp.TCPConn) {
	module := req.ReadString()
	rpcId := req.ReadUInt16()
	oldReqKey := req.GetReqKey()

	ids := meta.GetModuleIDs(module, meta.G_Local.Version)
	for _, id := range ids {
		if p := netConfig.GetTcpConn(module, id); p != nil {
			p.CallRpc(rpcId, func(buf *common.NetPack) {
				buf.WriteBuf(req.LeftBuf())
			}, func(backBuf *common.NetPack) {
				//异步回调，不能直接用ack
				backBuf.SetReqKey(oldReqKey)
				conn.WriteMsg(backBuf)
			})
		}
	}
}
