package logic

import (
	"fmt"
	"gamelog"
	"net/http"
	"netConfig"
	"tcp"

	"unsafe"
)

var (
	g_cache_battle_conn *tcp.TCPConn
)

func SendToBattle(msgID uint16, msgdata []byte) {
	if g_cache_battle_conn == nil {
		g_cache_battle_conn = netConfig.GetTcpConn("battle", 0)
	}
	g_cache_battle_conn.WriteMsg(msgID, msgdata)
}

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
	SendToBattle(1, buffer)
}

//////////////////////////////////////////////////////////////////////
//! 测试msg
func Hand_Msg_1(pTcpConn *tcp.TCPConn, buf []byte) {
	fmt.Println("Hand_Msg_1:")

	fmt.Println(*(*string)(unsafe.Pointer(&buf)))
}
