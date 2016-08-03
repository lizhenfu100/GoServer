package main

import (
	"http"
	// "tcp"
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

	http.Http_Client_Test_1()

	//主线程可以干别的去了
	time.Sleep(3 * time.Second)
}
