package http

import (
	"bytes"
	"common"
	"common/file"
	"errors"
	"gamelog"
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
	if ack, err := http.Post(url, "text/HTML", bytes.NewReader(b)); err == nil {
		//如果Response.Body既没有被完全读取，也没有被关闭，那么这次http事务就没有完成
		//除非连接因超时终止了，否则相关资源无法被回收
		return ReadResponse(ack)
	} else {
		gamelog.Error(err.Error())
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
	var err error
	if r.ContentLength > 0 { //http读大数据，r.ContentLength是-1
		ret = make([]byte, r.ContentLength)
		_, err = io.ReadFull(r.Body, ret)
	} else {
		ret, err = ioutil.ReadAll(r.Body)
	}
	if r.Body.Close(); err != nil {
		gamelog.Error("ReadBody: " + err.Error())
		return nil
	}
	return
}

// ------------------------------------------------------------
//! 模块注册
func RegistToSvr(destAddr string) {
	go func() {
		buf, _ := common.T2B(meta.G_Local)
		for {
			if PostReq(destAddr+"/reg_to_svr", buf) == nil {
				time.Sleep(3 * time.Second)
			} else {
				return
			}
		}
	}()
}

// ------------------------------------------------------------
//! 上传下载文件（大文件不适用，无断点续传）
func UploadFile(url, fullname string) error {
	fd, err := os.Open(fullname)
	if err != nil {
		return err
	}
	defer fd.Close()
	buf := &bytes.Buffer{}
	wr := multipart.NewWriter(buf)
	fw, err := wr.CreateFormFile("file", fullname)
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, fd); err != nil {
		return err
	}
	wr.Close()
	if resp, err := http.Post(url, wr.FormDataContentType(), buf); err == nil {
		resp.Body.Close()
		return nil
	} else {
		return err
	}
}
func DownloadFile(url, localDir, localName string) error {
	if buf := Get(url); buf != nil {
		if fd, err := file.CreateFile(localDir, localName, os.O_WRONLY|os.O_TRUNC); err == nil {
			fd.Write(buf)
			fd.Close()
			return nil
		} else {
			return err
		}
	} else {
		return ErrGet
	}
}
