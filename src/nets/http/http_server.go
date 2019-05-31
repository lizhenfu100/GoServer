package http

import (
	"common"
	"common/file"
	"conf"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"gamelog"
	"generate_out/err"
	"io"
	"netConfig/meta"
	"path"
	"svr_client/test/qps"
	"sync"
)

//TODO:zhoumf: 原生HttpHandle的参数与fasthttp如何统一？
//TODO:zhoumf: w http.ResponseWriter 可替换为 w io.Writer
//TODO:zhoumf: r *http.Request 关键是如何统一接口，能取到get/post参数

//idx1 := strings.Index(addr, "//") + 2
//idx2 := strings.LastIndex(addr, ":")
//ip := addr[idx1:idx2]
//port := common.Atoi(addr[idx2+1:])
func Addr(ip string, port uint16) string { return fmt.Sprintf("http://%s:%d", ip, port) }

func InitClient(c iClient) { Client = c }
func InitSvr(module string, svrId int) {
	if conf.TestFlag_CalcQPS {
		go qps.WatchLoop()
	}
	g_svraddr_path = fmt.Sprintf("%s/%s/%d/reg_addr.csv", file.GetExeDir(), module, svrId)
	loadCacheNetMeta()
}

// ------------------------------------------------------------
//! 模块注册
func Reg_to_svr(w io.Writer, req []byte) {
	errCode := err.Success //! 创建回复
	defer func() {
		ack := make([]byte, 2)
		binary.LittleEndian.PutUint16(ack, errCode)
		w.Write(ack)
	}()

	pMeta := new(meta.Meta)
	if e := common.B2T(req, pMeta); e != nil {
		errCode = err.Convert_err
		fmt.Println(e.Error())
		return
	}
	if p := meta.GetMeta(pMeta.Module, pMeta.SvrID); p != nil {
		if p.IP != pMeta.IP || p.OutIP != pMeta.OutIP {
			errCode = err.Data_repeat //防止配置错误，出现外网节点顶替
			fmt.Println("Error: RegistToSvr repeat: ", pMeta)
			return
		}
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
	if records, e := file.ReadCsv(g_svraddr_path); e == nil {
		metas := make([]meta.Meta, len(records))
		for i := 0; i < len(records); i++ {
			json.Unmarshal(common.S2B(records[i][0]), &metas[i])
			meta.AddMeta(&metas[i])
			fmt.Println("load meta cache: ", metas[i])
		}
	}
}
func appendNetMeta(pMeta *meta.Meta) {
	if meta.GetMeta("zookeeper", 0) != nil {
		return //有zookeeper实现重启恢复，不必本地缓存
	}
	_mutex.Lock()
	records, e := file.ReadCsv(g_svraddr_path)
	_mutex.Unlock()
	if e == nil {
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
	e = file.AppendCsv(dir, name, []string{common.B2S(b)})
	_mutex.Unlock()
	if e != nil {
		gamelog.Error("AppendSvrAddrCsv: " + e.Error())
	}
}
