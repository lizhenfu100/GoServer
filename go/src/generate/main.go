package main

import (
	"fmt"
	"runtime/debug"
	"time"
)

const (
	K_SvrDir = "../src/svr_"
	K_OutDir = "../src/generate_out/rpc/"
)

func main() {
	ptr := generatRpcRegist("center")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("cross")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("game")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("login")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("sdk")
	collectRpc_Go(ptr)
	ptr = generatRpcRegist("file")
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
