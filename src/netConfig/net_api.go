/***********************************************************************
* @ 选取节点id
* @ brief
	1、取模方式的模块，不支持动态增删
		、gateway、friend、分流game
		、center比较特殊，它本身是无状态的
			玩家分到错误节点也能重新从数据库读到自己数据

	2、无zookeeper的架构，只能重启扩容
		、节点间的meta信息，须保持一致性

* @ author zhoumf
* @ date 2019-1-18
***********************************************************************/
package netConfig

import (
	"common"
	"common/std/hash"
	"gamelog"
	"generate_out/rpc/enum"
	"math/rand"
	"netConfig/meta"
	"nets/http"
	"nets/tcp"
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
		n := hash.StrHash(key)
		return ids[n%uint32(length)]
	}
}
func SyncRelayToCenter(svrId int, rid uint16, req, ack *common.NetPack) {
	//【Notice：确保对center的调用是同步的】
	CallRpcCenter(svrId, rid, func(buf *common.NetPack) {
		buf.WriteBuf(req.LeftBuf())
	}, func(recvBuf *common.NetPack) {
		ack.WriteBuf(recvBuf.LeftBuf())
	})
}
func CallRpcCenter(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if addr := GetHttpAddr("center", svrId); addr != "" {
		http.CallRpc(addr, rid, sendFun, recvFun)
	} else {
		gamelog.Error("center nil: svrId(%d) rpcId(%d)", svrId, rid)
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
	svrId := HashGatewayID(accountId)
	if conn := GetTcpConn("gateway", svrId); conn != nil {
		conn.CallRpc(enum.Rpc_gateway_relay_player_msg, func(buf *common.NetPack) {
			buf.WriteUInt16(rid)
			buf.WriteUInt32(accountId)
			sendFun(buf)
		}, recvFun)
	} else {
		gamelog.Error("gateway nil: svrId(%d) rpcId(%d)", svrId, rid)
	}
}

// ------------------------------------------------------------
//! battle
var g_cache_battle_conn = make(map[int]*tcp.TCPConn)

func CallRpcBattle(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if conn := GetBattleConn(svrId); conn != nil {
		conn.CallRpc(rid, sendFun, recvFun)
	} else {
		gamelog.Error("battle nil: svrId(%d) rpcId(%d)", svrId, rid)
	}
}
func GetBattleConn(svrId int) *tcp.TCPConn {
	conn, _ := g_cache_battle_conn[svrId]
	if conn == nil || conn.IsClose() {
		conn = GetTcpConn("battle", svrId)
		g_cache_battle_conn[svrId] = conn
	}
	return conn
}

// ------------------------------------------------------------
//! game
var g_cache_game_conn = make(map[int]*tcp.TCPConn)

func CallRpcGame(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if conn := GetGameConn(svrId); conn != nil {
		conn.CallRpc(rid, sendFun, recvFun)
	} else {
		gamelog.Error("game nil: svrId(%d) rpcId(%d)", svrId, rid)
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
	svrId := HashFriendID(accountId)
	if addr := GetHttpAddr("friend", svrId); addr != "" {
		http.CallRpc(addr, rid, func(buf *common.NetPack) {
			buf.WriteUInt32(accountId)
			sendFun(buf)
		}, recvFun)
	} else {
		gamelog.Error("friend nil: svrId(%d) rpcId(%d)", svrId, rid)
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
	} else {
		gamelog.Error("cross nil: rpcId(%d)", rid)
	}
}

// ------------------------------------------------------------
//! sdk -- 单点
const kSdkSvrId = 0

func CallRpcSdk(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if addr := GetHttpAddr("sdk", kSdkSvrId); addr != "" {
		http.CallRpc(addr, rid, sendFun, recvFun)
	} else {
		gamelog.Error("sdk nil: svrId(%d) rpcId(%d)", kSdkSvrId, rid)
	}
}
