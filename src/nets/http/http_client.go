package http

import (
	"bytes"
	"common"
	"common/file"
	"encoding/binary"
	"errors"
	"gamelog"
	"generate_out/err"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"netConfig/meta"
	"os"
	"time"
)

//func init() {
//	http.DefaultClient.Timeout = 3 * time.Second
//}
var (
	ErrGet  = errors.New("http get failed")
	ErrPost = errors.New("http post failed")
)

// ------------------------------------------------------------
//! 底层接口，业务层一般用不到
func PostReq(url string, b []byte) []byte {
	if ack, e := http.Post(url, "text/HTML", bytes.NewReader(b)); e == nil {
		//如果Response.Body既没有被完全读取，也没有被关闭，那么这次http事务就没有完成
		//除非连接因超时终止了，否则相关资源无法被回收
		return ReadResponse(ack)
	} else {
		gamelog.Error(e.Error())
		return nil
	}
}
func Get(url string) []byte {
	if r, e := http.Get(url); e == nil {
		return ReadResponse(r)
	}
	return nil
}
func Post(url string, contentType string, b []byte) []byte {
	if r, e := http.Post(url, contentType, bytes.NewReader(b)); e == nil {
		return ReadResponse(r)
	}
	return nil
}
func PostForm(url string, data url.Values) []byte {
	if r, e := http.PostForm(url, data); e == nil {
		return ReadResponse(r)
	}
	return nil
}
func ReadResponse(r *http.Response) (ret []byte) {
	var e error
	if r.ContentLength > 0 { //http读大数据，r.ContentLength是-1
		ret = make([]byte, r.ContentLength)
		_, e = io.ReadFull(r.Body, ret)
	} else {
		ret, e = ioutil.ReadAll(r.Body)
	}
	if r.Body.Close(); e != nil {
		gamelog.Error("ReadBody: " + e.Error())
		return nil
	}
	return
}

// ------------------------------------------------------------
//! 模块注册
func RegistToSvr(destAddr string) {
	go func() {
		firstMsg, _ := common.T2B(meta.G_Local)
		for {
			if b := PostReq(destAddr+"/reg_to_svr", firstMsg); b == nil {
				time.Sleep(3 * time.Second)
			} else if e := binary.LittleEndian.Uint16(b); e != err.Success {
				panic("RegistToSvr fail")
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
	buf := &bytes.Buffer{}
	wr := multipart.NewWriter(buf)
	fw, e := wr.CreateFormFile("file", fullname)
	if e != nil {
		return e
	}
	if _, e = io.Copy(fw, fd); e != nil {
		return e
	}
	wr.Close()
	if resp, e := http.Post(url, wr.FormDataContentType(), buf); e == nil {
		resp.Body.Close()
		return nil
	} else {
		return e
	}
}
func DownloadFile(url, localDir, localName string) error {
	if buf := Get(url); buf != nil {
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
