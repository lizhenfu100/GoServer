/***********************************************************************
* @ https
* @ brief
	、CA私钥：openssl genrsa -out ca.key 1024
	、CA数字证书：openssl req -x509 -new -nodes -key ca.key -subj "/CN=tonybai.com" -out ca.crt -days 300
	、服务器私钥：openssl genrsa -out server.key 1024
	、服务器证书：
		openssl req -new -key server.key -subj "/CN=localhost" -out server.csr
		openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 300

* @ author zhoumf
* @ date 2019-7-17
***********************************************************************/
package https

import (
	"common"
	"fmt"
	"net/http"
	mhttp "nets/http"
	http2 "nets/http/http"
)

const (
	k_svr_crt = "rsa/server.crt" //服务端的数字证书文件路径
	k_svr_key = "rsa/server.key" //服务端的私钥文件路径
	k_ca_key  = "rsa/ca.key"
	k_ca_crt  = "rsa/ca.crt" //用于验证服务端证书
)

var _svr http.Server

func NewHttpServer(port uint16, module string, svrId int) error {
	mhttp.InitSvr(module, svrId)
	http.HandleFunc("/client_rpc", http2.HandleRpc)
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write(common.S2B("---------------- test ----------------"))
	})
	http.HandleFunc("/reg_to_svr", func(w http.ResponseWriter, r *http.Request) {
		mhttp.Reg_to_svr(w, http2.ReadBody(r.Body))
	})
	_svr.Addr = fmt.Sprintf(":%d", port)
	return _svr.ListenAndServeTLS(k_svr_crt, k_svr_key)
}
func CloseServer() { _svr.Close() }

// ------------------------------------------------------------
// -- rpc
func RegHandlePlayerRpc(cb func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/player_rpc", cb)
}
