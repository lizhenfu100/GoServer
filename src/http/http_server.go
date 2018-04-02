package http

import (
	"common"
	"common/file"
	"encoding/json"
	"fmt"
	"gamelog"
	"net/http"
	"netConfig/meta"
	"sync"
)

//idx1 := strings.Index(addr, "//") + 2
//idx2 := strings.LastIndex(addr, ":")
//ip := addr[idx1:idx2]
//port := common.CheckAtoiName(addr[idx2+1 : len(addr)-1])
func Addr(ip string, port uint16) string { return fmt.Sprintf("http://%s:%d/", ip, port) }

func NewHttpServer(addr string) error {
	LoadCacheNetMeta()
	http.HandleFunc("/reg_to_svr", _reg_to_svr)
	return http.ListenAndServe(addr, nil)
}

// ------------------------------------------------------------
//! 模块注册
func _reg_to_svr(w http.ResponseWriter, r *http.Request) {
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	pMeta := new(meta.Meta)
	if err := common.ToStruct(buffer, pMeta); err != nil {
		gamelog.Error("RegistToSvr common.ToStruct: %s", err.Error())
		return
	}
	gamelog.Info("RegistToSvr: {%s %d}", pMeta.Module, pMeta.SvrID)

	meta.AddMeta(pMeta)
	AppendNetMeta(pMeta)

	defer func() {
		w.Write([]byte("ok"))
	}()
}

// ------------------------------------------------------------
//! 本地存储“远程服务”注册地址，以备Local Http NetSvr崩溃重启
//! 不似tcp，对端不知道这边傻逼了( ▔___▔)y
//! 采用追加方式，同个“远程服务”的地址，会被最新追加的覆盖掉
var (
	g_svraddr_path = file.GetExeDir() + "reg_addr.csv"
	_mutex         sync.Mutex
)

func LoadCacheNetMeta() {
	records, err := file.ReadCsv(g_svraddr_path)
	if err != nil {
		return
	}
	pMeta := new(meta.Meta)
	for i := 0; i < len(records); i++ {
		json.Unmarshal([]byte(records[i][0]), pMeta)
		//Notice：之前可能有同个key的，被后面追加的覆盖
		meta.AddMeta(pMeta)
	}
	file.UpdateCsv(g_svraddr_path, [][]string{})
}
func AppendNetMeta(pMeta *meta.Meta) {
	if meta.GetMeta("zookeeper", 0) != nil {
		return //有zookeeper实现重启恢复，不必本地缓存
	}
	_mutex.Lock()
	defer _mutex.Unlock()
	b, _ := json.Marshal(pMeta)
	record := []string{string(b)}
	err := file.AppendCsv(g_svraddr_path, record)
	if err != nil {
		gamelog.Error("AppendSvrAddrCsv (%v): %s", record, err.Error())
	}
}
