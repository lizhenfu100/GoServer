package netConfig

import (
	"common"
	"net/http"
	"tcp"
)

type (
	TcpHandle        func(*tcp.TCPConn, *common.NetPack)
	HttpHandle       func(http.ResponseWriter, *http.Request)
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
	req := common.NewNetPackLen(int(r.ContentLength))
	r.Body.Read(req.DataPtr)

	//! 创建回复
	ack := common.NewNetPackCap(128)

	msgId := req.GetOpCode()
	pid := req.GetReqIdx()
	println("\nHttpMsg:", common.DebugRpcIdToName(msgId), "  playerId:", pid)

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
	}
}
