package http

import (
	"bytes"
	"common"
	"common/file"
	"gamelog"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"netConfig/meta"
	"os"
	"time"
)

//func init() {
//	http.DefaultClient.Timeout = 3 * time.Second
//}

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
func ReadResponse(r *http.Response) (ret []byte) {
	var err error
	if r.ContentLength > 0 { //http读大数据，r.ContentLength是-1
		ret = make([]byte, r.ContentLength)
		_, err = io.ReadFull(r.Body, ret)
	} else {
		ret, err = ioutil.ReadAll(r.Body)
	}
	if r.Body.Close(); err != nil {
		gamelog.Error("ReadBody: %s", err.Error())
		return nil
	}
	return
}

// ------------------------------------------------------------
//! 模块注册
func RegistToSvr(destAddr string, meta *meta.Meta) {
	go func() {
		buf, _ := common.ToBytes(meta)
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
	if res, err := http.Get(url); err == nil {
		defer res.Body.Close()
		if result, err := ioutil.ReadAll(res.Body); err == nil {
			if fd, err := file.CreateFile(localDir, localName, os.O_WRONLY|os.O_TRUNC); err == nil {
				fd.Write(result)
				fd.Close()
				return nil
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
}
