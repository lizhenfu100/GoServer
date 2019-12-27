package https

import (
	"bytes"
	"common"
	"crypto/tls"
	"crypto/x509"
	"gamelog"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	http2 "nets/http/http"
)

var (
	Client   client
	g_client *http.Client
)

type client struct{}

func (client) PostReq(url string, b []byte) []byte {
	if ack, e := g_client.Post(url, "application/octet-stream", bytes.NewReader(b)); e == nil {
		//如果Response.Body既没有被完全读取，也没有被关闭，那么这次http事务就没有完成
		//除非连接因超时终止了，否则相关资源无法被回收
		return http2.ReadBody(ack.Body)
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
	if r, e := g_client.Get(url); e == nil {
		return http2.ReadBody(r.Body)
	}
	return nil
}
func (client) Post(url string, contentType string, b []byte) []byte {
	if r, e := g_client.Post(url, contentType, bytes.NewReader(b)); e == nil {
		return http2.ReadBody(r.Body)
	}
	return nil
}
func (client) PostBody(url string, contentType string, body io.Reader) []byte {
	if r, e := g_client.Post(url, contentType, body); e == nil {
		return http2.ReadBody(r.Body)
	}
	return nil
}
func PostForm(url string, data url.Values) []byte {
	if r, e := g_client.PostForm(url, data); e == nil {
		return http2.ReadBody(r.Body)
	}
	return nil
}

// ------------------------------------------------------------
func init() {
	if caCrt, e := ioutil.ReadFile(k_ca_crt); e != nil {
		panic("ReadFile:" + e.Error())
	} else {
		cert := x509.NewCertPool()
		cert.AppendCertsFromPEM(caCrt)

		g_client = &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: cert},
		}}
	}
}
