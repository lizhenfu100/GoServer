package api

import (
	"encoding/json"
	"gamelog"
	"http"
	"msg"
	"netConfig"
	"tcp"
)

var (
	g_cache_game_conn *tcp.TCPConn
	g_cache_game_addr string
)

func SendToGame(msgID uint16, msgdata []byte) {
	if g_cache_game_conn == nil {
		g_cache_game_conn = netConfig.GetTcpConn("game", 0)
	}
	g_cache_game_conn.WriteMsg(msgID, msgdata)
}

func AddTempBattleSvr(gameAddr, selfAddr string, selfSvrID int) {
	var req msg.Msg_add_temp_svr_Req
	req.Addr = selfAddr
	req.SvrID = selfSvrID

	buf, _ := json.Marshal(&req)

	_, err := http.PostReq(gameAddr+"/add_temp_svr", buf)
	if err != nil {
		gamelog.Error("AddTempBattleSvr PostReq fail. Error: %s", err.Error())
	}
}
