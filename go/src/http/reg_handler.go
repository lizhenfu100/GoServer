package http

import (
	"common"
	"fmt"
	"net/http"
)

//Notice：http的消息处理，是另开goroutine调用的，所以函数中可阻塞；tcp就不行了
func RegHttpMsgHandler() {
	http.HandleFunc("/test_1", Hand_Test_1)
}

//////////////////////////////////////////////////////////////////////
//! 测试代码 gamesvr
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
		b, _ := common.ToBytes(&ack)
		w.Write(b)
	}()
}

//////////////////////////////////////////////////////////////////////
//! 测试代码 client
type MSG_Test_Req struct {
	PlayerID   int64
	SessionKey string
	Type       byte
}
type MSG_Test_Ack struct {
	RetCode byte
	Data    string
}

func Http_Client_Test_1() {
	reqUrl := "http://127.0.0.1:9002/test_1"
	buffer, err := PostMsg(reqUrl, &MSG_Test_Req{1, "zzz", 1})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var ack MSG_Test_Ack
	common.ToStruct(buffer, &ack)

	fmt.Println("MSG_Test_Ack:", ack)
}
