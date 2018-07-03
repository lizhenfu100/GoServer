package http

import (
	"bytes"
	"common"
	"gamelog"
	"io"
	"mime/multipart"
	"net/http"
	"netConfig/meta"
	"os"
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

// ------------------------------------------------------------
//! 上传下载文件
func UploadFile(url, filename string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fw, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		gamelog.Error("writing to buffer: %s", err.Error())
		return err
	}
	fh, err := os.Open(filename)
	if err != nil {
		gamelog.Error("opening file(%s): %s", filename, err.Error())
		return err
	}
	if _, err = io.Copy(fw, fh); err != nil {
		return err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	if resp, err := http.Post(url, contentType, bodyBuf); err == nil {
		resp.Body.Close()
		return nil
	} else {
		return err
	}
}
