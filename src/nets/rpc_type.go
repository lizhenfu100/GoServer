/***********************************************************************
* @ 原生rpc注册接口，比对player_rpc.go
* @ brief
	1、Tcp现在是：像c++一样，io线程只负责收发，数据交付给全局队列(G_RpcQueue)，主线程逐帧处理
		· 如有需要可：每条连接各自线程收数据，直接在io线程调注册的业务函数

	2、Http注册的响应函数，是另开goroutine调用的，须考虑多线程

	3、多线程架构，对强交互的业务不友好【见player.go】

* @ 系统缺陷
	1、没处理rpc失败，调用者也不知道rpc失败 …… 需要消息队列中间件

* @ author zhoumf
* @ date 2017-5-25
***********************************************************************/
package nets

import (
	"common"
	"net/http"
	"nets/rpc"
)

type (
	// 与player强绑定的rpc，见player_rpc.go
	// PlayerRpc func(req, ack *common.NetPack, this *Type)
	RpcFunc    func(req, ack *common.NetPack, conn common.Conn)
	HttpHandle http.HandlerFunc
)

func RegRpc(vs map[uint16]RpcFunc) {
	for k, v := range vs {
		rpc.G_HandleFunc[k] = v
	}
}
func RegHttpHandler(httpLst map[string]HttpHandle) {
	for k, v := range httpLst {
		http.HandleFunc(k, v)
	}
}
