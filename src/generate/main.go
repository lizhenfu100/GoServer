package main

import (
	"fmt"
	"runtime/debug"
	"time"
)

func main() {
	svrList := []string{
		"shared_svr/zookeeper",
		"shared_svr/svr_center",
		"shared_svr/svr_login",
		"shared_svr/svr_gateway",
		"shared_svr/svr_save",
		"shared_svr/svr_file",
		"shared_svr/svr_friend",
		"svr_cross",
		"svr_sdk",
		"svr_game",
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%v: %s", r, debug.Stack())
			time.Sleep(time.Minute)
		}
	}()

	//1、收集并注册RpcFunc -- 公共服务器、具体某游戏的业务服务器
	vec := make([]*RpcInfo, len(svrList))
	funcs := make([]string, 0, 1024) //小写开头
	for i, v := range svrList {
		vec[i] = generateRpcRegist(v)
	}
	for _, ptr := range vec {
		addRpc_Go(&funcs, ptr)
	}

	//2、收集并注册RpcFunc -- 战斗服、客户端
	addRpc_C(&funcs)  //svr_battle
	addRpc_CS(&funcs) //client

	//3、RpcFunc收集完毕，生成RpcEunm
	if generateRpcEnum(funcs) {
		//4、生成golang服务器的路由信息
		modules := []string{"client"}
		for _, ptr := range vec {
			modules = append(modules, ptr.Module)
		}
		generateRpcRoute(modules, funcs)
	}

	generateErrCode() //生成错误码

	print("Generate success...\n")
}
