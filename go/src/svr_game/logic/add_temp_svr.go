package logic

import (
	"encoding/json"
	"gamelog"
	"msg"
	"net/http"
	"netConfig"
	"svr_game/api"
	"tcp"
)

func Handle_Add_Temp_Svr(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	//! 解析消息
	var req msg.Msg_add_temp_svr_Req
	err := json.Unmarshal(buffer, &req)
	if err != nil {
		gamelog.Error("Handle_Add_Temp_Svr unmarshal fail. Error: %s", err.Error())
		return
	}

	client := &tcp.TCPClient{}
	client.OnConnected = func(conn *tcp.TCPConn) {
		api.AddBattleSvr(req.SvrID, conn)
	}
	client.ConnectToSvr(req.Addr, netConfig.G_Local_Module, netConfig.G_Local_SvrID)

	//! 创建回复
	defer func() {
		w.Write([]byte("ok"))
	}()
}
