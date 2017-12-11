package main

import (
	// "http"
	// "tcp"
	"fmt"
	// "netConfig"
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

	//test_reload_csv()
	lst := new([]int)
	lst1 := lst
	lst2 := lst
	*lst = append(*lst, 23)
	fmt.Println(lst, lst1, lst2)

	team := *lst
	*lst = append(team, 1)
	fmt.Println(lst, lst1, lst2)

	time.Sleep(100 * time.Second)
}
