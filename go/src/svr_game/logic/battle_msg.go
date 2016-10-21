package logic

import (
	"fmt"
	"gamelog"
	"net/http"
	"svr_game/api"
	"tcp"

	"unsafe"
)

//! 消息处理函数
//
func Handle_Battle_Echo(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	fmt.Println(*(*string)(unsafe.Pointer(&buffer)))

	//! 创建回复
	defer func() {
		w.Write([]byte("echo"))
	}()

	// 转发给Battle进程
	api.SendToBattle(1, 1, buffer)
}

//////////////////////////////////////////////////////////////////////
//! 测试msg
func Hand_Msg_1(pTcpConn *tcp.TCPConn, buf []byte) {
	fmt.Println("Hand_Msg_1:")

	fmt.Println(*(*string)(unsafe.Pointer(&buf)))
}
