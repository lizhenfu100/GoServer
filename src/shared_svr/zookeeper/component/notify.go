/***********************************************************************
* @ 供其它节点引用的zookeeper组件
* @ brief
    1、init()中须手动注册Rpc，代码生成器仅捕获模块自己目录下的

    2、每个节点连上zookeeper时，下发它要主动连接的节点，再通知要连接它的那些节点

* @ author zhoumf
* @ date 2018-3-13
***********************************************************************/
package component

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/http"
	"nets/tcp"
)

func init() {
	tcp.G_HandleFunc[enum.Rpc_node_join] = _Rpc_node_join1
	http.G_HandleFunc[enum.Rpc_node_join] = _Rpc_node_join
	tcp.G_HandleFunc[enum.Rpc_node_quit] = _Rpc_node_quit1
	http.G_HandleFunc[enum.Rpc_node_quit] = _Rpc_node_quit
}
func RegisterToZookeeper() {
	if pZoo := meta.GetMeta(meta.Zookeeper, 0); pZoo != nil {
		req, ack := func(buf *common.NetPack) {
			buf.WriteString(meta.G_Local.Module)
			buf.WriteInt(meta.G_Local.SvrID)
		}, func(recvBuf *common.NetPack) { //主动连接zoo通告的服务节点
			for len(recvBuf.LeftBuf()) > 0 {
				pMeta := new(meta.Meta) //Notice：须每次new新的，供ConnectToModule中的闭包引用
				pMeta.BufToData(recvBuf)
				netConfig.ConnectModule(pMeta)
			}
		}
		if pZoo.HttpPort > 0 {
			addr := http.Addr(pZoo.IP, pZoo.HttpPort)
			http.RegistToSvr(addr)
			http.CallRpc(addr, enum.Rpc_zoo_register, req, ack)
		} else {
			netConfig.ConnectModuleTcp(pZoo, func(conn *tcp.TCPConn) {
				conn.CallRpc(enum.Rpc_zoo_register, req, ack)
			})
		}
	}
}

//有服务节点加入，zoo通告相应客户节点
func _Rpc_node_join1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_node_join(req, ack) }
func _Rpc_node_join(req, ack *common.NetPack) {
	pMeta := new(meta.Meta)
	pMeta.BufToData(req)
	netConfig.ConnectModule(pMeta)
}
func _Rpc_node_quit1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_node_quit(req, ack) }
func _Rpc_node_quit(req, ack *common.NetPack) {
	module := req.ReadString()
	svrID := req.ReadInt()
	meta.DelMeta(module, svrID)
	//tcp node 消逝，由tcp系统自己感知，无需zookeeper额外处理
	//tcp client 会断线重连，目前tcp的DelMeta，仅在tcp_server调用
	//FIXME：用运维指令方式，主动剔除节点，阻断tcp_client的自动重连 -- 达到动态删除效果
}
