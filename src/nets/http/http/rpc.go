/***********************************************************************
* @ http rpc
* @ brief
	1、system rpc：将原生http的参数统一转换为NetPack
	2、player rpc：在system rpc基础之上，加了层find player逻辑，若找不到不处理

* @ Notic
	1、http的消息处理，是另开goroutine调用的，所以函数中可阻塞；tcp就不行了

	2、正因为每条消息都是另开goroutine，若玩家连续发多条消息，服务器就是并发处理了，存在竞态……client确保应答式通信

	3、http服务器自带多线程环境，写业务代码危险多了，须十分注意共享数据的保护
		· 全局变量
		· 队伍数据
		· 聊天记录（只要不是独属自己的数据，都得加保护~囧）

* @ http消息回调
	http._doRegistToSvr(0x8c8d60, 0xc042160380, 0xc0421c6000)
		D:/soulnet/GoServer/src/http/http_server.go:38 +0x3b
	net/http.HandlerFunc.ServeHTTP(0x76e638, 0x8c8d60, 0xc042160380, 0xc0421c6000)
		C:/Go/src/net/http/server.go:1918 +0x4b
	net/http.(*ServeMux).ServeHTTP(0x8fd800, 0x8c8d60, 0xc042160380, 0xc0421c6000)
		C:/Go/src/net/http/server.go:2254 +0x137
	net/http.serverHandler.ServeHTTP(0xc042158410, 0x8c8d60, 0xc042160380, 0xc0421c6000)
		C:/Go/src/net/http/server.go:2619 +0xbb
	net/http.(*conn).serve(0xc0421c0000, 0x8c91e0, 0xc0420343c0)
		C:/Go/src/net/http/server.go:1801 +0x724
	created by net/http.(*Server).Serve
		C:/Go/src/net/http/server.go:2720 +0x28f

* @ author zhoumf
* @ date 2017-8-10
***********************************************************************/
package http

import (
	"net/http"
	mhttp "nets/http"
	"strings"
)

func _HandleRpc(w http.ResponseWriter, r *http.Request) {
	//defer func() {//库已经有recover了，见net/http/server.go:1918
	//	if r := recover(); r != nil {
	//		gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
	//	}
	//	ack.Free()
	//}()
	if buf := ReadRequest(r); buf != nil {
		ip := ""
		if mhttp.G_Intercept != nil {
			ip = strings.Split(r.RemoteAddr, ":")[0]
		}
		mhttp.HandleRpc(buf, w, ip)
	}
}

// ------------------------------------------------------------
//! player rpc
func RegHandlePlayerRpc(cb func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/player_rpc", cb)
}
