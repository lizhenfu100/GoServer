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
	"net/http"
	"tcp"
)

type (
	HttpHandle       func(http.ResponseWriter, *http.Request)
	TcpHandle        func(*common.NetPack, *common.NetPack, *tcp.TCPConn)
	HttpPlayerHandle func(*common.NetPack, *common.NetPack, interface{})
)

var (
	g_http_player_handler = make(map[uint16]HttpPlayerHandle)

	//! 需要主动发给client的数据，每回通信时捎带过去
	G_Before_Recv_Player_Http func(uint32) interface{}
	G_After_Recv_Player_Http  func(interface{}, *common.NetPack)
)

func RegTcpHandler(tcpLst map[string]TcpHandle) {
	for k, v := range tcpLst {
		tcp.G_HandlerMsgMap[common.RpcNameToId(k)] = v
	}
}

//! 后台各系统间的数据传输格式可能有多种，比如Json，所以接口参数是原始http的
func RegHttpSystemHandler(httpLst map[string]HttpHandle) {
	for k, v := range httpLst {
		http.HandleFunc("/"+k, v)
	}
}
func RegHttpPlayerHandler(httpLst map[string]HttpPlayerHandle) {
	http.HandleFunc("/client_rpc", _HandleHttpPlayerMsg)

	for k, v := range httpLst {
		g_http_player_handler[common.RpcNameToId(k)] = v
	}
}
func _HandleHttpPlayerMsg(w http.ResponseWriter, r *http.Request) {
	//! 接收信息
	req := common.NewNetPackLen(int(r.ContentLength) - common.PACK_HEADER_SIZE)
	r.Body.Read(req.DataPtr)

	//! 创建回复
	ack := common.NewNetPackCap(128)

	msgId := req.GetOpCode()
	pid := req.GetReqIdx()
	println("\nHttpMsg:", common.DebugRpcIdToName(msgId), "len:", req.Size(), "  playerId:", pid)

	if handler, ok := g_http_player_handler[msgId]; ok {

		var player interface{}
		if G_Before_Recv_Player_Http != nil {
			player = G_Before_Recv_Player_Http(pid)
		}

		handler(req, ack, player)

		if G_After_Recv_Player_Http != nil && player != nil {
			G_After_Recv_Player_Http(player, ack)
		}

		if ack.BodySize() > 0 {
			w.Write(ack.DataPtr)
		}
	} else {
		println("\n===> HttpMsg:", common.DebugRpcIdToName(msgId), "Not Regist!!!")
	}
}
