/***********************************************************************
* @ 原生rpc注册接口，比对player_rpc.go
* @ brief
	1、Tcp现在是：像c++一样，io线程只负责收发，数据交付给全局队列(G_RpcQueue)，主线程逐帧处理
		· 如有需要可：每条连接各自线程收数据，直接在io线程调注册的业务函数

	2、Http注册的响应函数，是另开goroutine调用的，须考虑多线程

	3、多线程架构，对强交互的业务不友好【见player.go】

* @ author zhoumf
* @ date 2017-5-25
***********************************************************************/
package register

import (
	"common"
	"net/http"
	mhttp "nets/http"
	"nets/tcp"
)

type (
	// 与player强绑定的rpc，见player_rpc.go
	// PlayerRpc func(req, ack *common.NetPack, this *Type)
	TcpRpc     func(req, ack *common.NetPack, conn *tcp.TCPConn)
	HttpRpc    func(req, ack *common.NetPack)
	HttpHandle func(http.ResponseWriter, *http.Request)
)

func RegTcpRpc(tcpLst map[uint16]TcpRpc) {
	for k, v := range tcpLst {
		tcp.G_HandleFunc[k] = v
	}
}

//! 封装成NetPack的模块间通信；若需要其它传输格式(如Json)直接调http.HandleFunc(rpcname, func)注册
func RegHttpRpc(httpLst map[uint16]HttpRpc) {
	for k, v := range httpLst {
		mhttp.G_HandleFunc[k] = v
	}
}
func RegHttpHandler(httpLst map[string]HttpHandle) {
	for k, v := range httpLst {
		http.HandleFunc(k, v)
	}
}
