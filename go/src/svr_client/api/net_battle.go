package api

import (
	"netConfig"
	"tcp"
)

var (
	g_cache_battle_conn = make(map[int]*tcp.TCPConn)
)

func SendToBattle(svrID int, msgID uint16, msgdata []byte) {
    conn, _ := g_cache_battle_conn[svrID]
    if conn == nil {
        conn = netConfig.GetTcpConn("battle", svrID)
        g_cache_battle_conn[svrID] = conn
    }
    conn.WriteMsg(msgID, msgdata)
}
