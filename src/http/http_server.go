package http

import (
	"common"
	"common/net/meta"
	"encoding/json"
	"fmt"
	"gamelog"
	"net/http"
	"sync"
)

//idx1 := strings.Index(addr, "//") + 2
//idx2 := strings.LastIndex(addr, ":")
//ip := addr[idx1:idx2]
//port := common.CheckAtoiName(addr[idx2+1 : len(addr)-1])
func Addr(ip string, port uint16) string { return fmt.Sprintf("http://%s:%d/", ip, port) }

func NewHttpServer(addr string) error {
	LoadCacheNetMeta()
	http.HandleFunc("/reg_to_svr", _doRegistToSvr)
	return http.ListenAndServe(addr, nil)
}

// ------------------------------------------------------------
//! 模块注册
var g_reg_addr_map sync.Map

func _doRegistToSvr(w http.ResponseWriter, r *http.Request) {
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	ptr := new(meta.Meta)
	err := common.ToStruct(buffer, ptr)
	if err != nil {
		gamelog.Error("DoRegistToSvr common.ToStruct fail: %s", err.Error())
		return
	}
	gamelog.Debug("DoRegistToSvr: %v", *ptr)

	g_reg_addr_map.Store(common.KeyPair{ptr.Module, ptr.SvrID}, ptr)

	meta.AddMeta(ptr)
	AppendNetMeta(ptr)

	defer func() {
		w.Write([]byte("ok"))
	}()
}
func FindRegModuleAddr(module string, id int) string { //"http://%s:%d/"
	if v, ok := g_reg_addr_map.Load(common.KeyPair{module, id}); ok {
		ptr := v.(*meta.Meta)
		return Addr(ptr.IP, ptr.HttpPort)
	}
	gamelog.Error("FindRegModuleAddr nil: {%s:%d}", module, id)
	return ""
}
func GetRegModuleIDs(module string) (ret []int) {
	g_reg_addr_map.Range(func(k, v interface{}) bool {
		if k.(common.KeyPair).Name == module {
			ret = append(ret, k.(common.KeyPair).ID)
		}
		return true
	})
	return
}
func ForeachRegModule(f func(v *meta.Meta)) {
	g_reg_addr_map.Range(func(k, v interface{}) bool {
		f(v.(*meta.Meta))
		return true
	})
}

// ------------------------------------------------------------
//! 本地存储“远程服务”注册地址，以备Local Http NetSvr崩溃重启
//! 不似tcp，对端不知道这边傻逼了( ▔___▔)y
//! 采用追加方式，同个“远程服务”的地址，会被最新追加的覆盖掉
var (
	g_svraddr_path = common.GetExeDir() + "reg_addr.csv"
	_mutex         sync.Mutex
)

func LoadCacheNetMeta() {
	records, err := common.ReadCsv(g_svraddr_path)
	if err != nil {
		return
	}
	var info meta.Meta
	for i := 0; i < len(records); i++ {
		json.Unmarshal([]byte(records[i][0]), &info)
		//Notice：之前可能有同个key的，被后面追加的覆盖
		g_reg_addr_map.Store(common.KeyPair{info.Module, info.SvrID}, info)
	}
	common.UpdateCsv(g_svraddr_path, [][]string{})
}
func AppendNetMeta(meta *meta.Meta) {
	_mutex.Lock()
	defer _mutex.Unlock()
	b, _ := json.Marshal(meta)
	record := []string{string(b)}
	err := common.AppendCsv(g_svraddr_path, record)
	if err != nil {
		gamelog.Error("AppendSvrAddrCsv (%v): %s", record, err.Error())
	}
}
