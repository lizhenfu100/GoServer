package api

import (
	"common"
	"common/net/meta"
	"http"
	"netConfig"
	"sync"
)

var (
	g_cache_center_addr sync.Map // make(map[int]string)
	g_cache_game_addr   sync.Map // make(map[int]string)
)

// ------------------------------------------------------------
//! center
func CallRpcCenter(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	var addr string
	if v, ok := g_cache_center_addr.Load(svrId); ok {
		addr = v.(string)
	} else {
		addr = netConfig.GetHttpAddr("center", svrId)
		g_cache_center_addr.Store(svrId, addr)
	}
	http.CallRpc(addr, rid, sendFun, recvFun)
}

func SyncRelayToCenter(svrId int, rid uint16, req, ack *common.NetPack) {
	isSyncCall := false
	CallRpcCenter(svrId, rid, func(buf *common.NetPack) {
		buf.WriteBuf(req.Body())
	}, func(recvBuf *common.NetPack) {
		isSyncCall = true
		ack.WriteBuf(recvBuf.Body())
	})
	if isSyncCall == false {
		panic("Using ack int another CallRpc must be sync!!! zhoumf\n")
	}
}

func HashCenterID(key string) int {
	ids := meta.GetModuleIDs("center", netConfig.G_Local_Meta.Version)
	length := uint32(len(ids))
	n := common.StringHash(key)
	return ids[n%length]
}

// ------------------------------------------------------------
//! game
func CallRpcGame(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	var addr string
	if v, ok := g_cache_game_addr.Load(svrId); ok {
		addr = v.(string)
	} else {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr.Store(svrId, addr)
	}
	http.CallRpc(addr, rid, sendFun, recvFun)
}
