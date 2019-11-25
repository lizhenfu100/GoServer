package logic

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
)

func Rpc_gateway_relay_module(req *common.NetPack, recvFun func(*common.NetPack)) {
	rpcId := req.ReadUInt16()
	svrId := req.ReadInt()

	/*
		知道rpc属于哪个模块，但模块的具体路由方式不定呀~囧
		默认都按JumpHash，特殊的再调各接口？
	*/
	module := enum.GetRpcModule(rpcId)
	if svrId == -1 { //随机节点
		svrId = meta.RandModuleID(module)
	}
	if p, ok := netConfig.GetRpc(module, svrId); ok {
		p.CallRpcSafe(rpcId, func(buf *common.NetPack) {
			buf.WriteBuf(req.LeftBuf())
		}, recvFun)
	}
}
func Rpc_gateway_relay_modules(req *common.NetPack, recvFun func(*common.NetPack)) {
	rpcId := req.ReadUInt16()

	module := enum.GetRpcModule(rpcId)
	ids := meta.GetModuleIDs(module, meta.G_Local.Version)
	for _, id := range ids {
		if p, ok := netConfig.GetRpc(module, id); ok {
			p.CallRpcSafe(rpcId, func(buf *common.NetPack) {
				buf.WriteBuf(req.LeftBuf())
			}, recvFun)
		}
	}
}
