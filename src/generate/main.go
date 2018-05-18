package main

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

const (
	K_SvrDir = "../src/"
	K_OutDir = "../src/generate_out/"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%v: %s", r, debug.Stack())
			time.Sleep(time.Minute)
		}
	}()
	//1、收集并注册RpcFunc -- 公共服务器、具体某游戏的业务服务器
	funcs := make([]string, 0, 1024) //小写开头
	vec := []*RpcInfo{
		generatRpcRegist("shared_svr/zookeeper"),
		generatRpcRegist("shared_svr/svr_center"),
		generatRpcRegist("shared_svr/svr_login"),
		generatRpcRegist("shared_svr/svr_gateway"),
		generatRpcRegist("shared_svr/svr_save"),
		generatRpcRegist("shared_svr/svr_file"),
		generatRpcRegist("shared_svr/svr_friend"),
		generatRpcRegist("svr_cross"),
		generatRpcRegist("svr_sdk"),
		generatRpcRegist("svr_game"),
	}
	for _, ptr := range vec {
		addRpc_Go(&funcs, ptr)
	}

	//2、收集并注册RpcFunc -- 战斗服、客户端
	addRpc_C(&funcs)  //svr_battle
	addRpc_CS(&funcs) //client

	//3、RpcFunc收集完毕，生成RpcEunm
	if generatRpcEnum(funcs) {
		//4、生成golang服务器的路由信息
		modules := []string{"client"}
		for _, ptr := range vec {
			modules = append(modules, ptr.Module)
		}
		generatRpcRoute(modules, funcs)
	}
	fmt.Print("Generate success...\n")
}

func GetModuleName(svr string) string {
	if i := strings.LastIndex(svr, "/"); i >= 0 {
		svr = svr[i+1:]
	}
	if i := strings.Index(svr, "svr_"); i >= 0 {
		return svr[i+4:]
	}
	return svr
}
