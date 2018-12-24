package http

import (
	"common"
	"common/file"
	"encoding/json"
	"fmt"
	"gamelog"
	"io/ioutil"
	"net/http"
	"netConfig/meta"
	"path"
	"sync"
)

//idx1 := strings.Index(addr, "//") + 2
//idx2 := strings.LastIndex(addr, ":")
//ip := addr[idx1:idx2]
//port := common.CheckAtoiName(addr[idx2+1:])
func Addr(ip string, port uint16) string { return fmt.Sprintf("http://%s:%d", ip, port) }

var _svr http.Server

func NewHttpServer(port uint16, module string, svrId int) error {
	g_svraddr_path = fmt.Sprintf("%s/%s/%d/reg_addr.csv", file.GetExeDir(), module, svrId)
	loadCacheNetMeta()
	http.HandleFunc("/reg_to_svr", _reg_to_svr)
	_svr.Addr = fmt.Sprintf(":%d", port)
	return _svr.ListenAndServe()
}
func CloseServer() { _svr.Close() }

// ------------------------------------------------------------
//! 模块注册
func _reg_to_svr(w http.ResponseWriter, r *http.Request) {
	buf, _ := ioutil.ReadAll(r.Body)
	pMeta := new(meta.Meta)
	if err := common.ToStruct(buf, pMeta); err != nil {
		gamelog.Error("RegistToSvr common.ToStruct: %s", err.Error())
		return
	}

	meta.AddMeta(pMeta)
	appendNetMeta(pMeta)
	gamelog.Info("RegistToSvr: %v", pMeta)
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
		json.Unmarshal([]byte(records[i][0]), &metas[i])
		meta.AddMeta(&metas[i])
		gamelog.Info("loadCache: %v", &metas[i])
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
			json.Unmarshal([]byte(records[i][0]), pMeta2)
			if pMeta.IsSame(pMeta2) {
				return
			}
		}
	}

	b, _ := json.Marshal(pMeta)
	record := []string{string(b)}
	_mutex.Lock()
	dir, name := path.Split(g_svraddr_path)
	err = file.AppendCsv(dir, name, record)
	_mutex.Unlock()
	if err != nil {
		gamelog.Error("AppendSvrAddrCsv (%v): %s", record, err.Error())
	}
}
