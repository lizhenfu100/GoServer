/***********************************************************************
* @ 网络IO
* @ brief
	1、Tcp现在是：每条连接各自线程收数据，直接在io线程调注册的业务函数

	2、对强交互的业务不友好【见player.go】，需考虑多线程问题

	3、Http注册的响应函数，是另开goroutine调用的，也要考虑多线程

	4、如有需要，可像c++一样，io线程只负责收发，数据交付给全局队列，主线程逐帧处理，避免竞态

* @ author zhoumf
* @ date 2017-5-25
***********************************************************************/
package netConfig

import (
	"common"
	"http"
	nhttp "net/http"
	"tcp"
)

type (
	TcpRpc        func(req, ack *common.NetPack, conn *tcp.TCPConn)
	HttpHandle    func(w nhttp.ResponseWriter, r *nhttp.Request)
	HttpRpc       func(req, ack *common.NetPack)
	HttpPlayerRpc func(req, ack *common.NetPack, p interface{})
)

func RegTcpRpc(tcpLst map[uint16]TcpRpc) {
	for k, v := range tcpLst {
		tcp.G_HandlerMsgMap[k] = v
	}
}

//! 封装成NetPack的模块间通信；若需要其它传输格式(如Json)直接调http.HandleFunc(rpcname, func)注册
func RegHttpRpc(httpLst map[uint16]HttpRpc) {
	for k, v := range httpLst {
		http.G_HandlerMap[k] = v
	}
	http.RegHandleRpc()
}

//! 访问玩家数据的消息，要求该玩家已经登录，否则不处理
func RegHttpPlayerRpc(httpLst map[uint16]HttpPlayerRpc) {
	for k, v := range httpLst {
		http.G_PlayerHandlerMap[k] = v
	}
	http.RegHandlePlayerRpc()
}

func RegHttpHandler(httpLst map[string]HttpHandle) {
	for k, v := range httpLst {
		nhttp.HandleFunc("/"+k, v)
	}
}
