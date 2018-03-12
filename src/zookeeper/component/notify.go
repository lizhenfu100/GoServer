package component

import (
	"common"
	"common/net/meta"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"tcp"
)

func init() {
	tcp.G_HandleFunc[enum.Rpc_svr_node_join] = Rpc_svr_node_join
}

var (
	g_cache_zoo_conn *tcp.TCPConn
)

func CallRpcZoo(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if g_cache_zoo_conn == nil || g_cache_zoo_conn.IsClose() {
		g_cache_zoo_conn = netConfig.GetTcpConn("zookeeper", 0)
	}
	g_cache_zoo_conn.CallRpc(rid, sendFun, recvFun)
}

func RegisterToZookeeper() {
	// 初始化同zookeeper的连接，并注册
	if pMeta := meta.GetMeta("zookeeper", 0); pMeta != nil && g_cache_zoo_conn == nil {
		client := new(tcp.TCPClient)
		client.OnConnect = func(conn *tcp.TCPConn) {
			CallRpcZoo(enum.Rpc_zoo_register, func(buf *common.NetPack) {
				buf.WriteString(netConfig.G_Local_Meta.Module)
				buf.WriteInt(netConfig.G_Local_Meta.SvrID)
			}, func(recvBuf *common.NetPack) { //主动连接zoo通告的服务节点
				count := recvBuf.ReadInt()
				for i := 0; i < count; i++ {
					pMeta := new(meta.Meta) //Notice：须每次new新的
					pMeta.BufToData(recvBuf)
					ConnectToModule(pMeta)
				}
			})
		}
		client.ConnectToSvr(tcp.Addr(pMeta.IP, pMeta.TcpPort), netConfig.G_Local_Meta)
		netConfig.G_Client_Conns.Store(common.KeyPair{pMeta.Module, pMeta.SvrID}, client)
	}
}

//有服务节点加入，zoo通告相应客户节点
func Rpc_svr_node_join(req, ack *common.NetPack, conn *tcp.TCPConn) {
	pMeta := new(meta.Meta)
	pMeta.BufToData(req)
	ConnectToModule(pMeta)
}
func Rpc_http_node_quit(req, ack *common.NetPack, conn *tcp.TCPConn) {
	module := req.ReadString()
	svrID := req.ReadInt()
	meta.DelMeta(module, svrID)
	//tcp node 消逝，由tcp系统自己感知，无需zookeeper额外处理
	//tcp client 会断线重连，所以tcp的DelMeta，仅在tcp_server调用
	//FIXME：用运维指令方式，主动剔除节点，阻断tcp_client的自动重连 -- 达到动态删除效果
}

func ConnectToModule(ptr *meta.Meta) {
	if ptr.HttpPort > 0 {
		http.RegistToSvr(http.Addr(ptr.IP, ptr.HttpPort), netConfig.G_Local_Meta)
		meta.AddMeta(ptr)
	} else {
		client := new(tcp.TCPClient)
		client.OnConnect = func(conn *tcp.TCPConn) {
			netConfig.G_Client_Conns.Store(common.KeyPair{ptr.Module, ptr.SvrID}, client)
			meta.AddMeta(ptr) //Notice：闭包中引用外部指针，其内容可能变动，须额外注意
		}
		client.ConnectToSvr(tcp.Addr(ptr.IP, ptr.TcpPort), netConfig.G_Local_Meta)
	}
}
