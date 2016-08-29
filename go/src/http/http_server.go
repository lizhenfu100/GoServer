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

var g_reg_addr_list []Msg_Regist_To_HttpSvr

func DoRegistToSvr(w http.ResponseWriter, r *http.Request) {
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)

	var req Msg_Regist_To_HttpSvr
	err := common.ToStruct(buffer, &req)
	if err != nil {
		fmt.Println("DoRegistToSvr common.ToStruct fail. Error:", err.Error())
		return
	}

	fmt.Println(req)
	g_reg_addr_list = append(g_reg_addr_list, req)

	defer func() {
		w.Write([]byte("ok"))
	}()
}

func FindRegModuleAddr(module string, id int) string {
	max := len(g_reg_addr_list)
	for i := 0; i < max; i++ {
		data := &g_reg_addr_list[i]
		if data.Module == module && data.ID == id {
			return data.Addr
		}
	}
	return ""
}
