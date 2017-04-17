package logic

import (
	"common"
	// "fmt"
	"gamelog"
	"net/http"
	"svr_center/api"
)

func Rpc_GetGameSvrLst(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	cfgLst := api.GetGamesvrCfgLst()

	//! 创建回复\\\\\\\\\\\\\\\\\
	defer func() {
		backBuf := common.NewNetPackCap(64)
		backBuf.WriteByte(byte(len(cfgLst)))
		for _, v := range cfgLst {
			backBuf.WriteString(v.Module)
			backBuf.WriteInt(v.SvrID)
			backBuf.WriteString(v.OutIP)
			backBuf.WriteInt(v.HttpPort)
		}
		w.Write(backBuf.DataPtr)
	}()
}
