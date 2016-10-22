package http

import (
	"bytes"
	"common"
	"fmt"
	"net/http"
	"time"
)

func PostMsg(url string, pMsg interface{}) ([]byte, error) {
	b, _ := common.ToBytes(pMsg)
	resp, err := http.Post(url, "text/HTML", bytes.NewReader(b))
	if err == nil {
		backBuf := make([]byte, resp.ContentLength)
		resp.Body.Read(backBuf)
		resp.Body.Close()
		return backBuf, nil
	}
	return nil, err
}
func PostReq(url string, b []byte) ([]byte, error) {
	resp, err := http.Post(url, "text/HTML", bytes.NewReader(b))
	if err == nil {
		backBuf := make([]byte, resp.ContentLength)
		resp.Body.Read(backBuf)
		resp.Body.Close()
		return backBuf, nil
	}
	return nil, err
}

func RegistToSvr(destAddr, srcAddr, srcModule string, srcID int) {
	go _RegistToSvr(destAddr, srcAddr, srcModule, srcID)
}
func _RegistToSvr(destAddr, srcAddr, srcModule string, srcID int) {
	pMsg := &Msg_Regist_To_HttpSvr{srcAddr, srcModule, srcID}
	for {
		http.DefaultClient.Timeout = 2 * time.Second
		_, err := PostMsg(destAddr+"/reg_to_svr", pMsg)
		if err != nil {
			fmt.Printf("(%s) RegistToSvr failed: %s \n", srcModule, err.Error())
			time.Sleep(3 * time.Second)
		} else {
			return
		}
	}
}

type Msg_Regist_To_HttpSvr struct {
	Addr   string
	Module string
	ID     int
}
