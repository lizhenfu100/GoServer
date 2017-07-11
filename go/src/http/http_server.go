package http

import (
	"common"
	// "net"
	"fmt"
	"gamelog"
	"net/http"
	"strconv"
)

//Notice：http的消息处理，是另开goroutine调用的，所以函数中可阻塞；tcp就不行了
//Notice：正因为每条消息都是另开goroutine，若玩家连续发多条消息，服务器就是并发处理了，存在竞态……client确保应答式通信
func NewHttpServer(addr string) error {
	LoadSvrAddrCsv()

	http.HandleFunc("/reg_to_svr", _doRegistToSvr)

	return http.ListenAndServe(addr, nil)
	// listener, err := net.Listen("tcp", addr)
	// if err != nil {
	// 	return err
	// }
	// defer listener.Close()
	// return http.Serve(listener, nil)
}

//////////////////////////////////////////////////////////////////////
//! 模块注册
var g_reg_addr_map = make(map[common.KeyPair]string) //slice结构可能出现多次注册问题

func _doRegistToSvr(w http.ResponseWriter, r *http.Request) {
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	var req Msg_Regist_To_HttpSvr
	err := common.ToStruct(buffer, &req)
	if err != nil {
		fmt.Println("DoRegistToSvr common.ToStruct fail: ", err.Error())
		return
	}
	fmt.Println("DoRegistToSvr: ", req)

	oldAddr, ok := g_reg_addr_map[common.KeyPair{req.Module, req.ID}]
	if ok && oldAddr == req.Addr {
		return
	} else {
		g_reg_addr_map[common.KeyPair{req.Module, req.ID}] = req.Addr
		AppendSvrAddrCsv(req.Module, req.ID, req.Addr)
	}

	defer func() {
		w.Write([]byte("ok"))
	}()
}
func FindRegModuleAddr(module string, id int) string { //"http://%s:%d/"
	if v, ok := g_reg_addr_map[common.KeyPair{module, id}]; ok {
		return v
	}
	fmt.Println("FindRegModuleAddr nil: ", common.KeyPair{module, id}, g_reg_addr_map)
	return ""
}
func GetRegModuleIDs(module string) (ret []int) {
	for k, _ := range g_reg_addr_map {
		if k.Name == module {
			ret = append(ret, k.ID)
		}
	}
	return
}

//////////////////////////////////////////////////////////////////////
//! 本地存储“远程服务”注册地址，以备Local Http NetSvr崩溃重启

//! 不似tcp，对端不知道这边傻逼了( ▔___▔)y

//! 采用追加方式，同个“远程服务”的地址，会被最新追加的覆盖掉
var (
	g_svraddr_path = common.GetExePath() + "reg_addr.csv"
)

func LoadSvrAddrCsv() {
	records, err := common.LoadCsv(g_svraddr_path)
	if err != nil {
		return
	}

	var (
		module string
		svrID  int
	)
	for i := 0; i < len(records); i++ {
		module = records[i][0]
		svrID, _ = strconv.Atoi(records[i][1])
		//Notice：之前可能有同个key的，被后面追加的覆盖
		g_reg_addr_map[common.KeyPair{module, svrID}] = records[i][2]
	}
}
func AppendSvrAddrCsv(module string, id int, addr string) {
	record := []string{module, strconv.Itoa(id), addr}

	err := common.AppendCsv(g_svraddr_path, record)
	if err != nil {
		gamelog.Error("AppendSvrAddrCsv (%v): %s", record, err.Error())
	}
}
