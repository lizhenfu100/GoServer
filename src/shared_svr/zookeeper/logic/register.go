package logic

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig/meta"
	"nets/tcp"
)

func Rpc_zoo_register(req, ack *common.NetPack, conn *tcp.TCPConn) {
	//1、提取新节点信息
	module := req.ReadString()
	svrId := req.ReadInt()
	pMeta := meta.GetMeta(module, svrId)

	posInBuf, count := ack.BodySize(), uint32(0)
	ack.WriteUInt32(count)
	tcp.ForeachRegModule(func(v *tcp.TCPConn) {
		if ptr, ok := v.UserPtr.(*meta.Meta); ok {
			//2、回复要主动连接哪些节点
			if pMeta.IsMyServer(ptr) {
				ptr.DataToBuf(ack)
				count++
			}
			//3、通知要连接它的节点
			if pMeta.IsMyClient(ptr) {
				v.CallRpc(enum.Rpc_svr_node_join, func(buf *common.NetPack) {
					pMeta.DataToBuf(buf)
				}, nil)
			}
		}
	})
	ack.SetPos(posInBuf, count)
}
