/***********************************************************************
* @ HTTP
* @ brief
	1、非常不安全，恶意劫持路由节点，即可知道发往后台的数据，包括密码~

* @ http.Request
	r.RequestURI	除去域名或ip的url
		/backup_conf?passwd=&weekdays=&onlintlimit=&auto=&force=
	r.URL.RawQuery 	加密后的参数，不含?
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
	"io/ioutil"
	"net/http"
	mhttp "nets/http"
)

var _svr http.Server

func NewHttpServer(port uint16, module string, svrId int) error {
	mhttp.InitSvr(module, svrId)
	http.HandleFunc("/client_rpc", _HandleRpc)
	http.HandleFunc("/reg_to_svr", func(w http.ResponseWriter, r *http.Request) {
		mhttp.Reg_to_svr(w, ReadRequest(r))
	})
	_svr.Addr = fmt.Sprintf(":%d", port)
	return _svr.ListenAndServe()
}
func CloseServer() { _svr.Close() }

func ReadRequest(r *http.Request) []byte {
	buf, e := ioutil.ReadAll(r.Body)
	if r.Body.Close(); e != nil {
		gamelog.Error("ReadBody: " + e.Error())
		return nil
	}
	return buf
}
func ReadResponse(r *http.Response) []byte {
	buf, e := ioutil.ReadAll(r.Body)
	if r.Body.Close(); e != nil {
		gamelog.Error("ReadBody: " + e.Error())
		return nil
	}
	return buf
}
