package main

import (
	"fmt"
	"runtime/debug"
	"time"
)

const (
	K_SvrDir = "../src/"
	K_OutDir = "../src/generate_out/"
)

func main() {
	ptr := generatRpcRegist("svr_center")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("svr_cross")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("svr_game")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("svr_login")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("svr_sdk")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("svr_file")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("zookeeper")
	collectRpc_Go(ptr)

	collectRpc_C()
	collectRpc_CSharp()
	generatRpcEnum()

	if r := recover(); r != nil {
		fmt.Printf("%v: %s", r, debug.Stack())
		time.Sleep(time.Minute)
	} else {
		fmt.Println("Generate success...")
	}
}
