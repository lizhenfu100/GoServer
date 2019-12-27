package http

import (
	"bytes"
	"common"
	"gamelog"
	"io"
	"net/http"
	"net/url"
)

var Client client

type client struct{}

func (client) PostReq(url string, b []byte) []byte {
	if ack, e := http.Post(url, "application/octet-stream", bytes.NewReader(b)); e == nil {
		//如果Response.Body既没有被完全读取，也没有被关闭，那么这次http事务就没有完成
		//除非连接因超时终止了，否则相关资源无法被回收
		return ReadBody(ack.Body)
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
	if r, e := http.Get(url); e == nil {
		return ReadBody(r.Body)
	}
	return nil
}
func (client) Post(url string, contentType string, b []byte) []byte {
	if r, e := http.Post(url, contentType, bytes.NewReader(b)); e == nil {
		return ReadBody(r.Body)
	}
	return nil
}
func (client) PostBody(url string, contentType string, body io.Reader) []byte {
	if r, e := http.DefaultClient.Post(url, contentType, body); e == nil {
		return ReadBody(r.Body)
	}
	return nil
}
func PostForm(url string, data url.Values) []byte {
	if r, e := http.PostForm(url, data); e == nil {
		return ReadBody(r.Body)
	}
	return nil
}
