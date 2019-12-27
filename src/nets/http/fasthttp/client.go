package fasthttp

import (
	"common"
	"gamelog"
	http "github.com/valyala/fasthttp"
	"io"
)

var Client client

type client struct{}

func (client) PostReq(url string, b []byte) []byte {
	if _, body, e := http.Post(b, url, nil); e == nil {
		//如果Response.Body既没有被完全读取，也没有被关闭，那么这次http事务就没有完成
		//除非连接因超时终止了，否则相关资源无法被回收
		return body
	} else {
		if msg := common.NewNetPack(b); msg != nil {
			gamelog.Error("(%d) %s", msg.GetMsgId(), e.Error())
		} else {
			gamelog.Error(e.Error())
		}
		return nil
	}
}
func (client) Get(url string) []byte {
	if _, body, e := http.Get(nil, url); e == nil {
		return body
	}
	return nil
}
func (client) Post(url string, contentType string, b []byte) []byte {
	//TODO:设置contentType
	if _, body, e := http.Post(b, url, nil); e == nil {
		return body
	}
	return nil
}
func (client) PostForm(url string, args *http.Args) []byte {
	if _, body, e := http.Post(nil, url, args); e == nil {
		return body
	}
	return nil
}
func (client) PostBody(url string, contentType string, body io.Reader) []byte {
	//TODO:
	return nil
}
