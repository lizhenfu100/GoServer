package logic

import (
	"common"
	"common/net/meta"
	"generate_out/rpc/enum"
	"tcp"
)

//FIXME：两种构建网络通信的方式：
//一、每个节点连上来时，下发它要主动连接的节点，再通知要连接它的那些节点
//二、节点内部作缓存，参考api模块，无缓存时，http同步向zookeeper取 ---- “同步取”的方式可能影响业务性能

func Rpc_zoo_register(req, ack *common.NetPack, conn *tcp.TCPConn) {
	//1、提取新节点信息
	module := req.ReadString()
	svrId := req.ReadInt()
	pMeta := meta.GetMeta(module, svrId)

	posInBuf, count := ack.BodySize(), uint32(0)
	ack.WriteUInt32(count)
	tcp.ForeachRegModule(func(v tcp.TRegConn) {
		//2、回复要主动连接哪些节点
		if pMeta.IsMyServer(v.Meta) {
			v.Meta.DataToBuf(ack)
			count++
		}
		//3、通知要连接它的节点
		if pMeta.IsMyClient(v.Meta) {
			v.Conn.CallRpc(enum.Rpc_svr_node_join, func(buf *common.NetPack) {
				pMeta.DataToBuf(buf)
			}, nil)
		}
	})
	ack.SetPos(posInBuf, count)
}
