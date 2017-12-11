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
	pMeta := meta.GetMeta("zookeeper", -1)
	if g_cache_zoo_conn == nil && pMeta != nil {
		client := new(tcp.TCPClient)
		client.OnConnect = func(conn *tcp.TCPConn) {
			CallRpcZoo(enum.Rpc_zoo_register, func(buf *common.NetPack) {
				buf.WriteString(netConfig.G_Local_Meta.Module)
				buf.WriteInt(netConfig.G_Local_Meta.SvrID)
			}, func(recvBuf *common.NetPack) {
				pMeta := new(meta.Meta)
				count := recvBuf.ReadInt()
				for i := 0; i < count; i++ {
					pMeta.BufToData(recvBuf)
					ConnectToModule(pMeta)
				}
			})
		}
		client.ConnectToSvr(tcp.Addr(pMeta.IP, pMeta.TcpPort), netConfig.G_Local_Meta)
		netConfig.G_Client_Conns[common.KeyPair{pMeta.Module, pMeta.SvrID}] = client
	}
}
func Rpc_svr_node_join(req, ack *common.NetPack, conn *tcp.TCPConn) {
	pMeta := new(meta.Meta)
	pMeta.BufToData(req)
	ConnectToModule(pMeta)
}

/* Notice：
由于zookeeper只有一个，节点与zookeeper的仅一条连接
rpc中对 ConnectToModule() 的调用其实是单线程的
所以对 netConfig.G_Client_Conns 的操作无需锁
*/
func ConnectToModule(ptr *meta.Meta) {
	meta.AddMeta(ptr)
	if ptr.HttpPort > 0 {
		http.RegistToSvr(
			http.Addr(ptr.IP, ptr.HttpPort),
			netConfig.G_Local_Meta)
	} else {
		client := new(tcp.TCPClient)
		client.ConnectToSvr(
			tcp.Addr(ptr.IP, ptr.TcpPort),
			netConfig.G_Local_Meta)
		netConfig.G_Client_Conns[common.KeyPair{ptr.Module, ptr.SvrID}] = client
	}
}
