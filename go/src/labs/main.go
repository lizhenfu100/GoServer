package main

import (
	// "http"
	// "tcp"
	"common"
	"fmt"
	"netConfig"
	"time"
)

func main() {
	// client := tcp.TCPClient{}
	// client.ConnectToSvr("127.0.0.1:9001")
	// // msgdata := []byte{1, 2, 3}
	// msgdata := make([]byte, 100)
	// for {
	// 	if client.TcpConn != nil {
	// 		client.TcpConn.WriteMsg(2, msgdata)
	// 		// break
	// 	}
	// 	time.Sleep(200 * time.Millisecond)
	// }

	common.LoadAllCsv()
	for k, v := range common.G_MapCsv {
		fmt.Println(k, v)
	}

	//主线程可以干别的去了
	time.Sleep(3 * time.Second)
}
