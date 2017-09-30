package api

import (
	"common"
	"http"
	"netConfig"
)

var (
	g_cache_center_addr string
)

func CallRpcCenter(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if g_cache_center_addr == "" {
		g_cache_center_addr = netConfig.GetHttpAddr("center", -1)
	}
	http.CallRpc(g_cache_center_addr, rid, sendFun, recvFun)
}

func SyncRelayToCenter(rid uint16, req, ack *common.NetPack) {
	isSyncCall := false
	CallRpcCenter(rid, func(buf *common.NetPack) {
		buf.WriteBuf(req.Body())
	}, func(recvBuf *common.NetPack) {
		isSyncCall = true
		ack.WriteBuf(recvBuf.Body())
	})
	if isSyncCall == false {
		panic("Using ack int another CallRpc must be sync!!! zhoumf\n")
	}
}
