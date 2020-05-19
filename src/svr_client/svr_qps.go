package main

import (
	"common"
	"common/console"
	"fmt"
	"generate_out/rpc/enum"
	fasthttp2 "github.com/valyala/fasthttp"
	http2 "net/http"
	"net/url"
	"nets/http/fasthttp"
	"nets/http/http"
	"nets/tcp"
	"svr_client/test/qps"
)

// ------------------------------------------------------------
// qps: http fasthttp
func test0() {
	//const kURL = "http://192.168.1.111:7777/echo"
	const kURL = "http://120.78.152.152:7777/echo"
	q := make(url.Values)
	q.Set("txt", "zhoumf")
	for {
		http.PostForm(kURL, q)
	}
}
func test1() {
	http2.HandleFunc("/echo", func(w http2.ResponseWriter, r *http2.Request) {
		qps.AddQps()
	})
	http.NewServer(7777, false)
}
func test2() {
	fasthttp.HandleFunc("/echo", func(ctx *fasthttp2.RequestCtx) {
		qps.AddQps()
		//fmt.Println("body: ", common.B2S(ctx.Request.Body()))
		ctx.Write(ctx.Request.Body())
	})
	fasthttp.NewServer(7777, false)
}

// ------------------------------------------------------------
// qps: tcp
func tcp0() {
	var client tcp.TCPClient
	client.Connect("192.168.1.111:7777", func(conn *tcp.TCPConn) {
		fmt.Println("tcp qps begin ...")
		for {
			conn.CallEx(enum.Rpc_gm_cmd, func(buf *common.NetPack) {
			}, nil)
		}
	})
}
func tcp_svr() {
	console.Init()
	tcp.NewServer(7777, 5000, false)
}
