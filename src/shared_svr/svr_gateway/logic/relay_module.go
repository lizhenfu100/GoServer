package logic

import (
	"common"
	"generate_out/rpc/enum"
	"math/rand"
	"netConfig"
	"netConfig/meta"
	"nets/tcp"
)

func Rpc_gateway_relay_module(req, ack *common.NetPack, conn *tcp.TCPConn) {
	rpcId := req.ReadUInt16()
	svrId := req.ReadInt()
	oldReqKey := req.GetReqKey()

	/*
		知道rpc属于哪个模块，但模块的具体路由方式不定呀~囧
		默认都按JumpHash，特殊的再调各接口？
	*/
	module := enum.GetRpcModule(rpcId)
	if svrId == -1 { //随机节点
		ids := meta.GetModuleIDs(module, meta.G_Local.Version)
		if n := len(ids); n > 0 {
			svrId = ids[rand.Intn(n)]
		}
	}
	if p, ok := netConfig.GetRpc(module, svrId); ok {
		p.CallRpcSafe(rpcId, func(buf *common.NetPack) {
			buf.WriteBuf(req.LeftBuf())
		}, func(backBuf *common.NetPack) {
			//异步回调，不能直接用ack
			backBuf.SetReqKey(oldReqKey)
			conn.WriteMsg(backBuf)
		})
	}
}
func Rpc_gateway_relay_modules(req, ack *common.NetPack, conn *tcp.TCPConn) {
	rpcId := req.ReadUInt16()
	oldReqKey := req.GetReqKey()

	module := enum.GetRpcModule(rpcId)
	ids := meta.GetModuleIDs(module, meta.G_Local.Version)
	for _, id := range ids {
		if p, ok := netConfig.GetRpc(module, id); ok {
			p.CallRpcSafe(rpcId, func(buf *common.NetPack) {
				buf.WriteBuf(req.LeftBuf())
			}, func(backBuf *common.NetPack) {
				//异步回调，不能直接用ack
				backBuf.SetReqKey(oldReqKey)
				conn.WriteMsg(backBuf)
			})
		}
	}
}
