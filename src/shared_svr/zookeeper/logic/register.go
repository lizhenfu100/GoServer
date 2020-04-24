package logic

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/http"
	"nets/tcp"
)

func init() {
	tcp.G_HandleFunc[enum.Rpc_zoo_register] = _Rpc_zoo_register1
	http.G_HandleFunc[enum.Rpc_zoo_register] = _Rpc_zoo_register
	tcp.G_HandleFunc[enum.Rpc_net_error] = Rpc_net_error
}
func _Rpc_zoo_register1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_zoo_register(req, ack) }
func _Rpc_zoo_register(req, ack *common.NetPack) {
	module := req.ReadString()
	svrId := req.ReadInt()
	pNew := meta.GetMeta(module, svrId)
	meta.G_Metas.Range(func(_, v interface{}) bool {
		switch ptr := v.(*meta.Meta); pNew.IsMyServer(ptr) {
		case meta.CS: //新节点是此节点的client，回复节点meta
			ptr.DataToBuf(ack)
		case meta.SC: //新节点是此节点的server，通知此节点
			if p, ok := netConfig.GetRpc1(ptr); ok {
				p.CallRpcSafe(enum.Rpc_node_join, func(buf *common.NetPack) {
					pNew.DataToBuf(buf)
				}, nil)
			}
		}
		return true
	})
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
	//zookeeper通告http svr消逝
	if pDel, ok := conn.GetUser().(*meta.Meta); ok && pDel.HttpPort > 0 {
		OnMetaDel(pDel)
	}
}

// ------------------------------------------------------------
func OnMetaAdd(pNew *meta.Meta) {
	meta.G_Metas.Range(func(_, v interface{}) bool {
		if ptr := v.(*meta.Meta); pNew.IsMyServer(ptr) == meta.SC {
			//新节点是此节点的server，通知此节点
			if p, ok := netConfig.GetRpc1(ptr); ok {
				p.CallRpcSafe(enum.Rpc_node_join, func(buf *common.NetPack) {
					pNew.DataToBuf(buf)
				}, nil)
			}
		}
		return true
	})
}
func OnMetaDel(pDel *meta.Meta) {
	meta.G_Metas.Range(func(_, v interface{}) bool {
		if ptr := v.(*meta.Meta); pDel.IsMyServer(ptr) != meta.None {
			if p, ok := netConfig.GetRpc1(ptr); ok {
				p.CallRpcSafe(enum.Rpc_node_quit, func(buf *common.NetPack) {
					buf.WriteString(pDel.Module)
					buf.WriteInt(pDel.SvrID)
				}, nil)
			}
		}
		return true
	})
}
