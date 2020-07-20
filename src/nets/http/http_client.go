package http

import (
	"bytes"
	"common"
	"common/file"
	"encoding/binary"
	"fmt"
	"generate_out/err"
	"io"
	"mime/multipart"
	"netConfig/meta"
	"os"
	"time"
)

var (
	ErrGet  = common.Err("http get failed")
	ErrPost = common.Err("http post failed")

	Client iClient //指向不同的实现
)

type iClient interface {
	PostReq(url string, b []byte) []byte
	PostBody(url string, contentType string, body io.Reader) []byte
	Post(url string, contentType string, b []byte) []byte
	Get(url string) []byte
}

//idx1 := strings.Index(addr, "//") + 2
//idx2 := strings.LastIndex(addr, ":")
//ip := addr[idx1:idx2]
//port := common.Atoi(addr[idx2+1:])
func Addr(ip string, port uint16) string { return fmt.Sprintf("http://%s:%d", ip, port) }

// ------------------------------------------------------------
//! 模块注册
func RegistToSvr(destAddr string) {
	go func() {
		firstMsg, _ := common.T2B(meta.G_Local)
		for {
			if b := Client.PostReq(destAddr+"/reg_to_svr", firstMsg); b == nil {
				time.Sleep(3 * time.Second)
			} else if e := binary.LittleEndian.Uint16(b); e != err.Success {
				panic(fmt.Sprintf("RegistToSvr errcode: %d", e))
			} else {
				return
			}
		}
	}()
}

// ------------------------------------------------------------
//! 上传下载文件（大文件不适用，无断点续传）
func UploadFile(url, fullname string) error {
	fd, e := os.Open(fullname)
	if e != nil {
		return e
	}
	defer fd.Close()
	body := &bytes.Buffer{}
	wr := multipart.NewWriter(body)
	fw, e := wr.CreateFormFile("file", fullname)
	if e != nil {
		return e
	}
	if _, e = io.Copy(fw, fd); e != nil {
		return e
	}
	wr.Close()
	if Client.PostBody(url, wr.FormDataContentType(), body) == nil {
		return ErrPost
	}
	return nil
}
func DownloadFile(url, localDir, localName string) error {
	if buf := Client.Get(url); buf != nil {
		if fd, e := file.CreateFile(localDir, localName, os.O_WRONLY|os.O_TRUNC); e == nil {
			fd.Write(buf)
			fd.Close()
			return nil
		} else {
			return e
		}
	} else {
		return ErrGet
	}
}
