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
	HttpHandle func(*common.ByteBuffer, *common.ByteBuffer) interface{}
)

var (
	G_Tcp_Handler      map[string]TcpHandle  = nil
	G_Http_Handler     map[string]HttpHandle = nil
	G_Before_Recv_Http func(interface{}, *common.ByteBuffer)
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
	gamelog.Info("HttpMsg: %s", r.URL.String())

	key := strings.TrimLeft(r.URL.String(), "/")
	//! 接收信息
	req := common.NewByteBufferLen(int(r.ContentLength))
	r.Body.Read(req.DataPtr)

	//! 创建回复
	ack := common.NewByteBufferCap(64)

	if handler, ok := G_Http_Handler[key]; ok {

		player := handler(req, ack)

		if G_Before_Recv_Http != nil && player != nil {
			G_Before_Recv_Http(player, ack) //需要主动发给client的数据，每回通信时捎带过去
		}
		if ack.BodySize() > 0 {
			w.Write(ack.DataPtr)
		}
	}
}
