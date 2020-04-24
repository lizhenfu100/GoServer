/***********************************************************************
* @ HTTP
* @ brief
	1、非常不安全，恶意劫持路由节点，即可知道发往后台的数据，包括密码~
		· 登录消息，可以用非对称加密
		· 中途附带上token，防修改消息中的pid

* @ Notic
	1、http的消息处理，是另开goroutine调用的，所以函数中可阻塞；tcp就不行了

	2、正因为每条消息都是另开goroutine，若玩家连续发多条消息，服务器就是并发处理了，存在竞态……client确保应答式通信

	3、http服务器自带多线程环境，写业务代码危险多了，须十分注意共享数据的保护
		· 全局变量
		· 队伍数据
		· 聊天记录（只要不是独属自己的数据，都得加保护~囧）

* @ author zhoumf
* @ date 2019-3-18
***********************************************************************/
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

//idx1 := strings.Index(addr, "//") + 2
//idx2 := strings.LastIndex(addr, ":")
//ip := addr[idx1:idx2]
//port := common.Atoi(addr[idx2+1:])
func Addr(ip string, port uint16) string { return fmt.Sprintf("http://%s:%d", ip, port) }

func InitSvr(module string, svrId int) {
	if conf.TestFlag_CalcQPS {
		qps.Watch()
	}
	_cache_path = fmt.Sprintf("%s/%s/%d/reg_addr.csv", file.GetExeDir(), module, svrId)
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
		gamelog.Warn(e.Error())
		return
	}
	if p := meta.GetMeta(pMeta.Module, pMeta.SvrID); p != nil {
		if p.IP != pMeta.IP || p.OutIP != pMeta.OutIP {
			errCode = err.Data_repeat //防止配置错误，出现外网节点顶替
			gamelog.Warn("Regist repeat: %v", pMeta)
			return
		}
	}
	meta.AddMeta(pMeta)
	appendNetMeta(pMeta)
	gamelog.Debug("Regist: %v", pMeta)
}

// ------------------------------------------------------------
//! 本地存储“远程服务”注册地址，以备Local Http NetSvr崩溃重启
//! 不似tcp，对端不知道这边傻逼了( ▔___▔)y
//! 采用追加方式，同个“远程服务”的地址，会被最新追加的覆盖掉
var (
	_cache_path string
	_mutex      sync.Mutex
)

func loadCacheNetMeta() {
	if records, e := file.ReadCsv(_cache_path); e == nil {
		metas := make([]meta.Meta, len(records))
		for i := 0; i < len(records); i++ {
			json.Unmarshal(common.S2B(records[i][0]), &metas[i])
			meta.AddMeta(&metas[i])
			fmt.Println("load meta cache: ", metas[i])
		}
	}
}
func appendNetMeta(pMeta *meta.Meta) {
	if meta.GetMeta(meta.Zookeeper, 0) != nil {
		return //有zookeeper实现重启恢复，不必本地缓存
	}
	_mutex.Lock()
	records, e := file.ReadCsv(_cache_path)
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
	dir, name := path.Split(_cache_path)
	_mutex.Lock()
	e = file.AppendCsv(dir, name, []string{common.B2S(b)})
	_mutex.Unlock()
	if e != nil {
		gamelog.Error("AppendSvrAddrCsv: " + e.Error())
	}
}
