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
		return body
	} else {
		if msg := common.ToNetPack(b); msg != nil {
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
