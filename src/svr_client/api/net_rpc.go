package api

import (
	"common"
	"netConfig"
	"tcp"
)

// ------------------------------------------------------------
//! battle
var (
	g_cache_battle_conn = make(map[int]*tcp.TCPConn)
)

func CallRpcBattle(svrID int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if conn := GetBattleConn(svrID); conn != nil {
		conn.CallRpc(rid, sendFun, recvFun)
	}
}
func GetBattleConn(svrID int) *tcp.TCPConn {
	conn, _ := g_cache_battle_conn[svrID]
	if conn == nil || conn.IsClose() {
		conn = netConfig.GetTcpConn("battle", svrID)
		g_cache_battle_conn[svrID] = conn
	}
	return conn
}
