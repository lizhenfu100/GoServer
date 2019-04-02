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

* @ author zhoumf
* @ date 2019-3-18
***********************************************************************/
package http

import (
	"common"
	"common/file"
	"conf"
	"encoding/json"
	"fmt"
	"gamelog"
	"io"
	"io/ioutil"
	"net/http"
	"netConfig/meta"
	"path"
	"svr_client/test/qps"
	"sync"
)

//idx1 := strings.Index(addr, "//") + 2
//idx2 := strings.LastIndex(addr, ":")
//ip := addr[idx1:idx2]
//port := common.Atoi(addr[idx2+1:])
func Addr(ip string, port uint16) string { return fmt.Sprintf("http://%s:%d", ip, port) }

var _svr http.Server

func NewHttpServer(port uint16, module string, svrId int) error {
	if conf.TestFlag_CalcQPS {
		go qps.WatchLoop()
	}
	g_svraddr_path = fmt.Sprintf("%s/%s/%d/reg_addr.csv", file.GetExeDir(), module, svrId)
	loadCacheNetMeta()
	http.HandleFunc("/reg_to_svr", _reg_to_svr)
	http.HandleFunc("/client_rpc", _HandleRpc)
	_svr.Addr = fmt.Sprintf(":%d", port)
	return _svr.ListenAndServe()
}
func CloseServer() { _svr.Close() }

func ReadRequest(r *http.Request) (req *common.NetPack) {
	var err error
	var buf []byte
	if r.ContentLength > 0 { //http读大数据，r.ContentLength是-1
		buf = make([]byte, r.ContentLength)
		_, err = io.ReadFull(r.Body, buf)
	} else {
		buf, err = ioutil.ReadAll(r.Body)
	}
	if r.Body.Close(); err != nil {
		gamelog.Error("ReadBody: " + err.Error())
		return nil
	}
	return common.NewNetPack(buf)
}

// ------------------------------------------------------------
//! 模块注册
func _reg_to_svr(w http.ResponseWriter, r *http.Request) {
	buf, _ := ioutil.ReadAll(r.Body)
	pMeta := new(meta.Meta)
	if err := common.B2T(buf, pMeta); err != nil {
		gamelog.Error("common.B2T: " + err.Error())
		return
	}
	meta.AddMeta(pMeta)
	appendNetMeta(pMeta)
	fmt.Println("RegistToSvr: ", pMeta)
}

// ------------------------------------------------------------
//! 本地存储“远程服务”注册地址，以备Local Http NetSvr崩溃重启
//! 不似tcp，对端不知道这边傻逼了( ▔___▔)y
//! 采用追加方式，同个“远程服务”的地址，会被最新追加的覆盖掉
var (
	g_svraddr_path string
	_mutex         sync.Mutex
)

func loadCacheNetMeta() {
	records, err := file.ReadCsv(g_svraddr_path)
	if err != nil {
		return
	}
	metas := make([]meta.Meta, len(records))
	for i := 0; i < len(records); i++ {
		json.Unmarshal(common.S2B(records[i][0]), &metas[i])
		meta.AddMeta(&metas[i])
		fmt.Println("load meta cache: ", metas[i])
	}
}
func appendNetMeta(pMeta *meta.Meta) {
	if meta.GetMeta("zookeeper", 0) != nil {
		return //有zookeeper实现重启恢复，不必本地缓存
	}
	_mutex.Lock()
	records, err := file.ReadCsv(g_svraddr_path)
	_mutex.Unlock()
	if err == nil {
		pMeta2 := new(meta.Meta)
		for i := 0; i < len(records); i++ {
			json.Unmarshal(common.S2B(records[i][0]), pMeta2)
			if pMeta.IsSame(pMeta2) {
				return
			}
		}
	}
	b, _ := json.Marshal(pMeta)
	dir, name := path.Split(g_svraddr_path)
	_mutex.Lock()
	err = file.AppendCsv(dir, name, []string{common.B2S(b)})
	_mutex.Unlock()
	if err != nil {
		gamelog.Error("AppendSvrAddrCsv: " + err.Error())
	}
}
