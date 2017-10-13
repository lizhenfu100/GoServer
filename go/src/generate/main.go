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
	generatRpcEnum()
	generatRpcRegist("center")
	generatRpcRegist("cross")
	generatRpcRegist("game")
	generatRpcRegist("login")
	generatRpcRegist("sdk")

	if r := recover(); r != nil {
		fmt.Printf("%v: %s", r, debug.Stack())
		time.Sleep(time.Minute)
	} else {
		fmt.Println("Generate success...")
	}
}
