package http

import (
	"bytes"
	"common"
	"gamelog"
	"io"
	"io/ioutil"
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
		//如果Response.Body既没有被完全读取，也没有被关闭，那么这次http事务就没有完成
		//除非连接因超时终止了，否则相关资源无法被回收
		ret := ReadResponse(ack)
		ack.Body.Close()
		return ret
	} else {
		gamelog.Error("PostReq url: %s \r\nerr: %s \r\n", url, err.Error())
		return nil
	}
}
func ReadResponse(r *http.Response) (ret []byte) {
	var err error
	if r.ContentLength > 0 {
		ret = make([]byte, r.ContentLength)
		_, err = io.ReadFull(r.Body, ret)
	} else {
		ret, err = ioutil.ReadAll(r.Body)
	}
	if err != nil {
		gamelog.Error("ReadBody: %s", err.Error())
		return nil
	}
	return
}

// ------------------------------------------------------------
//! 模块注册
func RegistToSvr(destAddr string, meta *meta.Meta) {
	go _registToSvr(destAddr, meta)
}
func _registToSvr(destAddr string, meta *meta.Meta) {
	buf, _ := common.ToBytes(meta)
	for {
		if PostReq(destAddr+"/reg_to_svr", buf) == nil {
			time.Sleep(3 * time.Second)
		} else {
			return
		}
	}
}

// ------------------------------------------------------------
//! 上传下载文件
func UploadFile(url, filename string) error {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		gamelog.Error("writing to buffer: %s", err.Error())
		return err
	}
	fd, err := os.Open(filename)
	if err != nil {
		gamelog.Error("opening file(%s): %s", filename, err.Error())
		return err
	}
	defer fd.Close()
	if _, err = io.Copy(fw, fd); err != nil {
		gamelog.Error("io.Copy: %s: %s", filename, err.Error())
		return err
	}
	contentType := w.FormDataContentType()
	w.Close()
	if resp, err := http.Post(url, contentType, buf); err == nil {
		resp.Body.Close()
		return nil
	} else {
		gamelog.Error("http.Post: %s", err.Error())
		return err
	}
}
