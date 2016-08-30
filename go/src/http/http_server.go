package http

import (
	"common"
	// "net"
	"fmt"
	"net/http"
)

func NewHttpServer(addr string) error {
	http.HandleFunc("/reg_to_svr", DoRegistToSvr)

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
type HttpAddrKey struct {
	Name string
	ID   int
}

var g_reg_addr_map = make(map[HttpAddrKey]string) //slice结构可能出现多次注册问题

func DoRegistToSvr(w http.ResponseWriter, r *http.Request) {
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	var req Msg_Regist_To_HttpSvr
	err := common.ToStruct(buffer, &req)
	if err != nil {
		fmt.Println("DoRegistToSvr common.ToStruct fail: ", err.Error())
		return
	}

	fmt.Println("DoRegistToSvr: ", req)
	g_reg_addr_map[HttpAddrKey{req.Module, req.ID}] = req.Addr

	defer func() {
		w.Write([]byte("ok"))
	}()
}

func FindRegModuleAddr(module string, id int) string {
	if v, ok := g_reg_addr_map[HttpAddrKey{module, id}]; ok {
		return v
	}
	fmt.Println("FindRegModuleAddr nil: ", HttpAddrKey{module, id}, g_reg_addr_map)
	return ""
}
