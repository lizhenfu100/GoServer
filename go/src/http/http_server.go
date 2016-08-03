package http

import (
	// "net"
	"net/http"
)

func NewHttpServer(addr string) error {
	return http.ListenAndServe(addr, nil)
	// listener, err := net.Listen("tcp", addr)
	// if err != nil {
	// 	return err
	// }
	// defer listener.Close()
	// return http.Serve(listener, nil)
}
