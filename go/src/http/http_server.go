package http

import (
	"common"
	"fmt"
	"net"
	"net/http"
)

func NewHttpServer(addr string, max int) error {
	if max <= 0 {
		return http.ListenAndServe(addr, nil)
	} else {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		defer listener.Close()
		return http.Serve(listener, nil)
	}
	return nil
}
func RegHttpMsgHandler() {
	http.HandleFunc("/test_1", Hand_Test_1)
}

//////////////////////////////////////////////////////////////////////
//! 测试msg
//////////////////////////////////////////////////////////////////////
func Hand_Test_1(w http.ResponseWriter, r *http.Request) {
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	var req MSG_Test_Req
	err := common.ToStruct(buffer, &req)
	if err != nil {
		fmt.Println("Hand_Test_1 bytes fail. Error:", err.Error())
		return
	}

	fmt.Println("MSG_Test_Req:", req)

	var ack MSG_Test_Ack
	ack.RetCode = 111
	ack.Data = "aaaaaa"
	defer func() {
		b, _ := common.ToBytes(ack)
		w.Write(b)
	}()
}
