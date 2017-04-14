package logic

import (
	"common"
	"fmt"
	"gamelog"
	"net/http"
	"svr_game/api"
)

//! 消息处理函数
//
func Handle_Client2Battle_Echo(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	msg := common.NewNetPackLen(int(r.ContentLength))
	r.Body.Read(msg.DataPtr)
	fmt.Println(msg.ReadString())

	//! 创建回复
	defer func() {
		w.Write([]byte("echo"))
	}()

	// 转发给Battle进程
	// api.SendToBattle(1, msg)
	api.SendToCross(msg)
}
