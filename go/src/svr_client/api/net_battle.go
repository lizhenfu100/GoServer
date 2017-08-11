package api

import (
	"common"
	"netConfig"
	"tcp"
)

var (
	g_cache_battle_conn = make(map[int]*tcp.TCPConn)
)

func CallRpcBattle(svrID int, rpc string, sendFun, recvFun func(*common.NetPack)) {
	GetBattleConn(svrID).CallRpc(rpc, sendFun, recvFun)
}
func GetBattleConn(svrID int) *tcp.TCPConn {
	conn, _ := g_cache_battle_conn[svrID]
	if conn == nil || conn.IsClose() {
		conn = netConfig.GetTcpConn("battle", svrID)
		g_cache_battle_conn[svrID] = conn
	}
	return conn
}
