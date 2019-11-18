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
	pNew := meta.GetMeta(module, svrId)

	posInBuf, count := ack.BodySize(), uint32(0)
	ack.WriteUInt32(count)
	tcp.ForeachRegModule(func(v *tcp.TCPConn) {
		if ptr, ok := v.UserPtr.(*meta.Meta); ok {
			switch pNew.IsMyServer(ptr) {
			case meta.CS: //2、新节点是此节点的client，回复节点meta
				ptr.DataToBuf(ack)
				count++
			case meta.SC: //3、新节点是此节点的server，通知此节点
				v.CallRpc(enum.Rpc_svr_node_join, func(buf *common.NetPack) {
					pNew.DataToBuf(buf)
				}, nil)
			}
		}
	})
	ack.SetPos(posInBuf, count)
}
