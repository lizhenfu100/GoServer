package main

import (
	"common"
	"common/console"
	"generate_out/rpc/enum"
	fasthttp2 "github.com/valyala/fasthttp"
	http2 "net/http"
	"net/url"
	"nets/http/fasthttp"
	"nets/http/http"
	"nets/tcp"
	"svr_client/test/qps"
)

func testQPS() {
	tcp1()
}

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
	go http.NewHttpServer(7777)
}
func test2() {
	fasthttp.HandleFunc("/echo", func(ctx *fasthttp2.RequestCtx) {
		qps.AddQps()
		//fmt.Println("body: ", common.B2S(ctx.Request.Body()))
		ctx.Write(ctx.Request.Body())
	})
	go fasthttp.NewHttpServer(7777)
}

// ------------------------------------------------------------
// qps: tcp
func tcp0() {
	var client tcp.TCPClient
	client.ConnectToSvr("192.168.1.111:7777", func(conn *tcp.TCPConn) {
		for {
			conn.CallRpcSafe(enum.Rpc_gm_cmd, func(buf *common.NetPack) {
			}, nil)
		}
	})
}
func tcp1() {
	console.Init()
	go tcp.NewTcpServer(7777, 5000)
}
