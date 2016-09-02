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
