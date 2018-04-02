package http

import (
	"bytes"
	"common"
	"gamelog"
	"net/http"
	"netConfig/meta"
	"time"
)

func init() {
	http.DefaultClient.Timeout = 3 * time.Second
}

// ------------------------------------------------------------
//! 底层接口，业务层一般用不到
func PostReq(url string, b []byte) []byte {
	if ack, err := http.Post(url, "text/HTML", bytes.NewReader(b)); err == nil {
		backBuf := make([]byte, ack.ContentLength)
		ack.Body.Read(backBuf)
		ack.Body.Close()
		return backBuf
	} else {
		gamelog.Error("PostReq url: %s \r\nerr: %s \r\n", url, err.Error())
		return nil
	}
}

// ------------------------------------------------------------
//! 模块注册
func RegistToSvr(destAddr string, meta *meta.Meta) {
	go _registToSvr(destAddr, meta)
}
func _registToSvr(destAddr string, meta *meta.Meta) {
	buf, _ := common.ToBytes(meta)
	for {
		if PostReq(destAddr+"reg_to_svr", buf) == nil {
			time.Sleep(3 * time.Second)
		} else {
			return
		}
	}
}
