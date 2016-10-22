package api

import (
	"netConfig"
	"tcp"
)

//Notice：TCPConn是对真正net.Conn的包装，发生断线重连时，会执行tcp.TCPConn.ResetConn()，所以外部缓存的tcp.TCPConn仍有效，无需更新
var (
	g_cache_battle_conn  = make(map[int]*tcp.TCPConn)
	g_cache_battle_svrid []int
)

func SendToBattle(svrID int, msgID uint16, msgdata []byte) {
	conn, _ := g_cache_battle_conn[svrID]
	if conn == nil {
		conn = netConfig.GetTcpConn("battle", svrID)
		g_cache_battle_conn[svrID] = conn
	}
	conn.WriteMsg(msgID, msgdata)
}

func InitBattleSvrID() {
	for i := 0; i < len(netConfig.G_SvrNetCfg); i++ {
		cfg := &netConfig.G_SvrNetCfg[i]
		if cfg.Module == "battle" {
			g_cache_battle_svrid = append(g_cache_battle_svrid, cfg.SvrID)
		}
	}
}
func AddBattleSvr(svrID int, conn *tcp.TCPConn) {
	g_cache_battle_svrid = append(g_cache_battle_svrid, svrID)
	g_cache_battle_conn[svrID] = conn
}

//Notice：玩家登录时调用，将svrID存到player struct，避免临时新增服务器时，分配计算结果不同
func AllocBattleSvrID(playerID int32) int {
	idx := int(playerID) % len(g_cache_battle_svrid)
	return g_cache_battle_svrid[idx]
}
