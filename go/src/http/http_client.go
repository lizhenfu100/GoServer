package http

import (
	"bytes"
	"common"
	"net/http"
)

func PostMsg(url string, pMsg interface{}) ([]byte, error) {
	b, _ := common.ToBytes(pMsg)
	resp, err := http.Post(url, "text/HTML", bytes.NewReader(b))
	backBuf := make([]byte, resp.ContentLength)
	resp.Body.Read(backBuf)
	resp.Body.Close()
	return backBuf, err
}
