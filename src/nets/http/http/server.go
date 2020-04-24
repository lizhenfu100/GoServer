/***********************************************************************
* @ HTTP
* @ http.Request
	r.RequestURI 除去域名或ip的url
		/backup_conf?passwd=&weekdays=&onlintlimit=&auto=&force=
	r.URL.RawQuery 加密后的参数，不含?
		passwd=&weekdays=&onlintlimit=&auto=&force=
	r.URL.Path
		/backup_conf

* @ http relay
	u, _ := url.Parse(newAddr + r.RequestURI) //除去域名或ip的url
	if buf := http.Client.Get(u.String()); buf != nil {
		w.Write(buf)
	}

* @ author zhoumf
* @ date 2019-3-18
***********************************************************************/
package http

import (
	"fmt"
	"gamelog"
	"io"
	"io/ioutil"
	"net/http"
	mhttp "nets/http"
	"strings"
)

func init() {
	// 默认用原生http实现底层通信，可替换成https、fasthttp
	mhttp.Client = Client
}

var _svr http.Server

func NewHttpServer(port uint16, block bool) {
	http.HandleFunc("/client_rpc", HandleRpc)
	http.HandleFunc("/reg_to_svr", func(w http.ResponseWriter, r *http.Request) {
		mhttp.Reg_to_svr(w, ReadBody(r.Body))
	})
	if _svr.Addr = fmt.Sprintf(":%d", port); block {
		_svr.ListenAndServe()
	} else {
		go _svr.ListenAndServe()
	}
}
func CloseServer() { _svr.Close() }

func ReadBody(body io.ReadCloser) []byte {
	buf, e := ioutil.ReadAll(body)
	if body.Close(); e != nil {
		gamelog.Error("ReadBody: " + e.Error())
		return nil
	}
	return buf
}

// ------------------------------------------------------------
// -- rpc
func HandleRpc(w http.ResponseWriter, r *http.Request) {
	//defer func() {//库已经有recover了，见net/http/server.go:1918
	//	if r := recover(); r != nil {
	//		gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
	//	}
	//	ack.Free()
	//}()
	if buf := ReadBody(r.Body); buf != nil {
		ip := ""
		if mhttp.Intercept() != nil {
			ip = strings.Split(r.RemoteAddr, ":")[0]
		}
		mhttp.HandleRpc(buf, w, ip)
	}
}
func RegHandlePlayerRpc(cb func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/player_rpc", cb)
}
