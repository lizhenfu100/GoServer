package api

import (
	"common"
	"netConfig"
	"tcp"
)

//Notice：TCPConn是对真正net.Conn的包装，发生断线重连时，会执行tcp.TCPConn.ResetConn()，所以外部缓存的tcp.TCPConn仍有效，无需更新
var (
	g_cache_battle_conn = make(map[int]*tcp.TCPConn)
	g_cache_game_conn   = make(map[int]*tcp.TCPConn)
)

// ------------------------------------------------------------
//! battle
func CallRpcBattle(svrID int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	GetBattleConn(svrID).CallRpc(rid, sendFun, recvFun)
}
func GetBattleConn(svrID int) *tcp.TCPConn {
	conn, _ := g_cache_battle_conn[svrID]
	if conn == nil || conn.IsClose() {
		conn = netConfig.GetTcpConn("battle", svrID)
		g_cache_battle_conn[svrID] = conn
	}
	return conn
}

// ------------------------------------------------------------
//! game
func CallRpcGame(svrID int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	GetGameConn(svrID).CallRpc(rid, sendFun, recvFun)
}
func GetGameConn(svrId int) *tcp.TCPConn {
	conn, _ := g_cache_game_conn[svrId]
	if conn == nil || conn.IsClose() {
		conn = netConfig.GetTcpConn("game", svrId)
		g_cache_game_conn[svrId] = conn
	}
	return conn
}
