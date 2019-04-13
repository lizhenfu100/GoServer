package fasthttp

import (
	"common"
	"encoding/binary"
	"errors"
	"gamelog"
	"generate_out/err"
	http "github.com/valyala/fasthttp"
	"netConfig/meta"
	"time"
)

var (
	ErrGet  = errors.New("fasthttp get failed")
	ErrPost = errors.New("fasthttp post failed")
)

// ------------------------------------------------------------
//! 底层接口，业务层一般用不到
func PostReq(url string, b []byte) []byte {
	if _, body, e := http.Post(b, url, nil); e == nil {
		//如果Response.Body既没有被完全读取，也没有被关闭，那么这次http事务就没有完成
		//除非连接因超时终止了，否则相关资源无法被回收
		return body
	} else {
		gamelog.Error(e.Error())
		return nil
	}
}
func Get(url string) []byte {
	if _, body, e := http.Get(nil, url); e == nil {
		return body
	}
	return nil
}
func PostForm(url string, args *http.Args) []byte {
	if _, body, e := http.Post(nil, url, args); e == nil {
		return body
	}
	return nil
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
