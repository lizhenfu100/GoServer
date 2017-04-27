package netConfig

import (
	"common"
	"gamelog"
	"net/http"
	"strings"
	"tcp"
)

type (
	TcpHandle  func(*tcp.TCPConn, *common.NetPack)
	HttpHandle func(*common.NetPack, *common.NetPack, interface{})
)

var (
	G_Tcp_Handler  map[string]TcpHandle  = nil
	G_Http_Handler map[string]HttpHandle = nil

	//! 需要主动发给client的数据，每回通信时捎带过去
	G_Before_Recv_Http func(uint32, *common.NetPack) interface{}
)

func RegMsgHandler() {
	if G_Http_Handler != nil {
		for k, _ := range G_Http_Handler {
			http.HandleFunc("/"+k, _HandleHttpMsg)
		}
	}
	if G_Tcp_Handler != nil {
		for k, v := range G_Tcp_Handler {
			tcp.G_HandlerMsgMap[common.RpcNameToId(k)] = v
		}
	}
}
func _HandleHttpMsg(w http.ResponseWriter, r *http.Request) {
	//! 接收信息
	req := common.NewNetPackLen(int(r.ContentLength))
	r.Body.Read(req.DataPtr)

	//! 创建回复
	ack := common.NewNetPackCap(64)

	//! http通信包头没用，拿来填充playerId
	pid := req.GetReqIdx()
	key := strings.TrimLeft(r.URL.String(), "/")
	gamelog.Info("HttpMsg: %s, pid(%d)", key, pid)

	if handler, ok := G_Http_Handler[key]; ok {
		if G_Before_Recv_Http != nil {
			player := G_Before_Recv_Http(pid, ack)
			handler(req, ack, player)
		} else {
			handler(req, ack, nil)
		}
		if ack.BodySize() > 0 {
			w.Write(ack.DataPtr)
		}
	}
}
