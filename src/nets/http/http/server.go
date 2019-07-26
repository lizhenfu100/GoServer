/***********************************************************************
* @ HTTP
* @ brief
	1、非常不安全，恶意劫持路由节点，即可知道发往后台的数据，包括密码~

* @ Notic
	1、http的消息处理，是另开goroutine调用的，所以函数中可阻塞；tcp就不行了

	2、正因为每条消息都是另开goroutine，若玩家连续发多条消息，服务器就是并发处理了，存在竞态……client确保应答式通信

	3、http服务器自带多线程环境，写业务代码危险多了，须十分注意共享数据的保护
		· 全局变量
		· 队伍数据
		· 聊天记录（只要不是独属自己的数据，都得加保护~囧）

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
	"io"
	"io/ioutil"
	"net/http"
	mhttp "nets/http"
	"strings"
)

var _svr http.Server

func NewHttpServer(port uint16, module string, svrId int) error {
	mhttp.InitSvr(module, svrId)
	http.HandleFunc("/client_rpc", HandleRpc)
	http.HandleFunc("/reg_to_svr", func(w http.ResponseWriter, r *http.Request) {
		mhttp.Reg_to_svr(w, ReadBody(r.Body))
	})
	_svr.Addr = fmt.Sprintf(":%d", port)
	return _svr.ListenAndServe()
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
		if mhttp.G_Intercept != nil {
			ip = strings.Split(r.RemoteAddr, ":")[0]
		}
		mhttp.HandleRpc(buf, w, ip)
	}
}
func RegHandlePlayerRpc(cb func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/player_rpc", cb)
}
