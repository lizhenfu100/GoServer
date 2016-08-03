package http

import (
	"bytes"
	"common"
	"fmt"
	"net/http"
)

func PostMsg(url string, msg interface{}) {
	b, _ := common.ToBytes(msg)
	_, err := PostServerReq(url, b)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
func PostServerReq(url string, buf []byte) ([]byte, error) {
	resp, err := http.Post(url, "text/HTML", bytes.NewReader(buf))
	buffer := make([]byte, resp.ContentLength)
	resp.Body.Read(buffer)
	resp.Body.Close()
	return buffer, err
}

//////////////////////////////////////////////////////////////////////
//! 测试msg
//////////////////////////////////////////////////////////////////////
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
	req := MSG_Test_Req{1, "zzz", 1}
	bytes, _ := common.ToBytes(req)
	buffer, err := PostServerReq(reqUrl, bytes)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var ack MSG_Test_Ack
	common.ToStruct(buffer, &ack)

	fmt.Println("MSG_Test_Ack:", ack)
}
