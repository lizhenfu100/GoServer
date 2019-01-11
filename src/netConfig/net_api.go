package netConfig

import (
	"common"
	"generate_out/rpc/enum"
	"http"
	"math/rand"
	"netConfig/meta"
	"tcp"
)

// ------------------------------------------------------------
//! center -- 账号名hash取模
func HashCenterID(key string) int {
	ids := meta.GetModuleIDs("center", meta.G_Local.Version)
	if length := len(ids); length <= 0 {
		return -1
	} else if length == 1 {
		return ids[0]
	} else {
		n := common.StringHash(key)
		return ids[n%uint32(length)]
	}
}
func SyncRelayToCenter(svrId int, rid uint16, req, ack *common.NetPack) {
	isSyncCall := false
	CallRpcCenter(svrId, rid, func(buf *common.NetPack) {
		buf.WriteBuf(req.LeftBuf())
	}, func(recvBuf *common.NetPack) {
		isSyncCall = true
		ack.WriteBuf(recvBuf.LeftBuf())
	})
	if isSyncCall == false {
		panic("Using ack int another CallRpc must be sync!!! zhoumf\n")
	}
}
func CallRpcCenter(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if addr := GetHttpAddr("center", svrId); addr != "" {
		http.CallRpc(addr, rid, sendFun, recvFun)
	}
}

// ------------------------------------------------------------
//! gateway -- 账号hash取模
var g_cache_gate_ids []int

func HashGatewayID(accountId uint32) int { //FIXME：考虑用一致性hash，取模方式导致gateway无法动态扩展
	length := uint32(len(g_cache_gate_ids))
	if length == 0 {
		g_cache_gate_ids = meta.GetModuleIDs("gateway", meta.G_Local.Version)
		length = uint32(len(g_cache_gate_ids))
	}
	if length > 0 {
		return g_cache_gate_ids[accountId%length]
	}
	return -1
}
func CallRpcGateway(accountId uint32, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if conn := GetTcpConn("gateway", HashGatewayID(accountId)); conn != nil {
		conn.CallRpc(enum.Rpc_gateway_relay_player_msg, func(buf *common.NetPack) {
			buf.WriteUInt16(rid)
			buf.WriteUInt32(accountId)
			sendFun(buf)
		}, recvFun)
	}
}

// ------------------------------------------------------------
//! battle
var g_cache_battle_conn = make(map[int]*tcp.TCPConn)

func CallRpcBattle(svrID int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if conn := GetBattleConn(svrID); conn != nil {
		conn.CallRpc(rid, sendFun, recvFun)
	}
}
func GetBattleConn(svrID int) *tcp.TCPConn {
	conn, _ := g_cache_battle_conn[svrID]
	if conn == nil || conn.IsClose() {
		conn = GetTcpConn("battle", svrID)
		g_cache_battle_conn[svrID] = conn
	}
	return conn
}

// ------------------------------------------------------------
//! game
var g_cache_game_conn = make(map[int]*tcp.TCPConn)

func CallRpcGame(svrID int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if conn := GetGameConn(svrID); conn != nil {
		conn.CallRpc(rid, sendFun, recvFun)
	}
}
func GetGameConn(svrId int) *tcp.TCPConn {
	conn, _ := g_cache_game_conn[svrId]
	if conn == nil || conn.IsClose() {
		conn = GetTcpConn("game", svrId)
		g_cache_game_conn[svrId] = conn
	}
	return conn
}

// ------------------------------------------------------------
//! friend -- 账号hash取模
func HashFriendID(accountId uint32) int {
	ids := meta.GetModuleIDs("friend", meta.G_Local.Version)
	if length := uint32(len(ids)); length > 0 {
		return ids[accountId%length]
	}
	return -1
}
func CallRpcFriend(accountId uint32, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if addr := GetHttpAddr("friend", HashFriendID(accountId)); addr != "" {
		http.CallRpc(addr, rid, func(buf *common.NetPack) {
			buf.WriteUInt32(accountId)
			sendFun(buf)
		}, recvFun)
	}
}

// ------------------------------------------------------------
//! cross -- 随机节点
func CallRpcCross(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	ids := meta.GetModuleIDs("cross", meta.G_Local.Version)
	if length := len(ids); length > 0 {
		id := ids[rand.Intn(length)]
		if conn := GetTcpConn("cross", id); conn != nil {
			conn.CallRpc(rid, sendFun, recvFun)
		}
	}
}

// ------------------------------------------------------------
//! sdk -- 单点
const kSdkSvrId = 0

func CallRpcSdk(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if addr := GetHttpAddr("sdk", kSdkSvrId); addr != "" {
		http.CallRpc(addr, rid, sendFun, recvFun)
	}
}
