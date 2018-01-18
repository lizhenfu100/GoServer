package http

import (
	"bytes"
	"common"
	"common/net/meta"
	"gamelog"
	"net/http"
	"time"
)

func init() {
	http.DefaultClient.Timeout = 3 * time.Second
}

// ------------------------------------------------------------
//! 底层接口，业务层一般用不到
func PostReq(url string, b []byte) []byte {
	ack, err := http.Post(url, "text/HTML", bytes.NewReader(b))
	if err == nil {
		backBuf := make([]byte, ack.ContentLength)
		ack.Body.Read(backBuf)
		ack.Body.Close()
		return backBuf
	} else {
		gamelog.Error("PostReq url: %s \r\nerr: %s \r\n", url, err.Error())
		return nil
	}
}

// 已验证：此函数失败，resp是nil，那resp.Body.Close()就不能无脑调了
// resp, err := http.Post(url, "text/HTML", bytes.NewReader(b))
// resp.Body.Close()

// ------------------------------------------------------------
//! 模块注册
func RegistToSvr(destAddr string, meta *meta.Meta) {
	go _RegistToSvr(destAddr, meta)
}
func _RegistToSvr(destAddr string, meta *meta.Meta) {
	buf, _ := common.ToBytes(meta)
	for {
		if PostReq(destAddr+"reg_to_svr", buf) == nil {
			time.Sleep(3 * time.Second)
		} else {
			return
		}
	}
}
