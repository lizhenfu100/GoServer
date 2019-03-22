package fasthttp

import (
	"common"
	"errors"
	"gamelog"
	http "github.com/valyala/fasthttp"
	"netConfig/meta"
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
	if _, body, err := http.Post(b, url, nil); err == nil {
		//如果Response.Body既没有被完全读取，也没有被关闭，那么这次http事务就没有完成
		//除非连接因超时终止了，否则相关资源无法被回收
		return body
	} else {
		gamelog.Error(err.Error())
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
func RegistToSvr(destAddr string, meta *meta.Meta) {
	go func() {
		buf, _ := common.T2B(meta)
		for {
			if PostReq(destAddr+"/reg_to_svr", buf) == nil {
				time.Sleep(3 * time.Second)
			} else {
				return
			}
		}
	}()
}
