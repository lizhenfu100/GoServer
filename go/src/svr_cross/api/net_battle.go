package api

import (
	"common"
	"netConfig"
	"tcp"
)

//Notice：TCPConn是对真正net.Conn的包装，发生断线重连时，会执行tcp.TCPConn.ResetConn()，所以外部缓存的tcp.TCPConn仍有效，无需更新
var (
	g_cache_battle_conn  = make(map[int]*tcp.TCPConn)
	g_cache_battle_svrid []int
)

func SendToBattle(svrID int, msg *common.NetPack) {
	conn, _ := g_cache_battle_conn[svrID]
	if conn == nil {
		conn = netConfig.GetTcpConn("battle", svrID)
		g_cache_battle_conn[svrID] = conn
	}
	conn.WriteMsg(msg)
}
func AddBattleSvr(svrID int, conn *tcp.TCPConn) {
	g_cache_battle_svrid = append(g_cache_battle_svrid, svrID)
	g_cache_battle_conn[svrID] = conn
}
func AllocBattleSvrID(playerID int32) int {
	idx := int(playerID) % len(g_cache_battle_svrid)
	return g_cache_battle_svrid[idx]
}
