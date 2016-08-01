package main

import (
	"fmt"
	"gamelog"
	"tcp"
	"time"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("labs", true)
	gamelog.SetLevel(0)

	client := tcp.TCPClient{}
	client.ConnectToSvr("127.0.0.1:9001")

	// msgdata := []byte{1, 2, 3}
	// for {
	// 	if client.TcpConn != nil {
	// 		client.TcpConn.WriteMsg(2, msgdata)
	// 		break
	// 	}
	// 	time.Sleep(1 * time.Second)
	// }

	//主线程可以干别的去了
	fmt.Println("--- WriteMsg Over")
	time.Sleep(3000 * time.Second)
}
