package api

import (
	"common"
	"common/net/meta"
	"http"
	"netConfig"
	"strings"
)

var (
	g_cache_center_addr = make(map[int]string)
	g_cache_game_addr   = make(map[int]string)
)

// ------------------------------------------------------------
//! center
func CallRpcCenter(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	addr, ok := g_cache_center_addr[svrId]
	if false == ok {
		addr = netConfig.GetHttpAddr("center", svrId)
		g_cache_center_addr[svrId] = addr
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
	ids := meta.GetModuleIDs("center")
	length := uint32(len(ids))
	n := common.StringHash(key)
	return ids[n%length]
}

// ------------------------------------------------------------
//! game
func CallRpcGame(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	addr, ok := g_cache_game_addr[svrId]
	if false == ok {
		addr = netConfig.GetHttpAddr("game", svrId)
		g_cache_game_addr[svrId] = addr
	}
	http.CallRpc(addr, rid, sendFun, recvFun)
}

func WriteRegGamesvr(buf *common.NetPack) {
	ids := meta.GetModuleIDs("game")
	buf.WriteByte(byte(len(ids)))
	for _, id := range ids {
		addr := netConfig.GetHttpAddr("game", id)

		idx1 := strings.Index(addr, "//") + 2
		idx2 := strings.LastIndex(addr, ":")
		ip := addr[idx1:idx2]
		port := common.CheckAtoiName(addr[idx2+1 : len(addr)-1])

		buf.WriteInt(id)
		buf.WriteString("ChillyRoom")
		buf.WriteString(ip)
		buf.WriteUInt16(uint16(port))
	}
}
