package main

import (
	"flag"
	"fmt"
	"runtime/debug"
	"sort"
	"time"
)

const (
	// 源码目录
	K_SvrDir = "../src/"
	// 生成文件的根目录
	K_OutDir = "../src/generate_out/"

	// 生成rpc枚举
	K_EnumOutDir   = K_OutDir + "rpc/enum/"
	K_EnumFileName = "generate_rpc_enum"

	// 生成rpc注册文件
	K_RegistOutDir   = K_OutDir + "rpc/"
	K_RegistFileName = "generate_rpc.go"

	// 生成error枚举
	K_ErrOutDir   = K_OutDir + "err/"
	K_ErrFileName = "generate_err_code"
)

var (
	// c++、c# rpc函数所在文件
	K_RpcFuncFile_C  = "../../CXServer/src/rpc/RpcEnum.h"
	K_RpcFuncFile_CS = "../../GameClient/Assets/RGScript/Net/Player/Player.cs"

	// c++、c# rpc枚举的输出目录
	K_EnumOutDir_C  = "../../CXServer/src/rpc/"
	K_EnumOutDir_CS = "../../GameClient/Assets/RGScript/generate/"

	// c++、c# errCode的输出目录
	K_ErrOutDir_C  = "../../CXServer/src/common/generate/"
	K_ErrOutDir_CS = "../../GameClient/Assets/RGScript/generate/"
)

func init() {
	flag.StringVar(&K_RpcFuncFile_C, "RpcFunc1", K_RpcFuncFile_C, "c++ Rpc函数文件")
	flag.StringVar(&K_RpcFuncFile_CS, "RpcFunc2", K_RpcFuncFile_CS, "c# Rpc函数文件")
	flag.StringVar(&K_EnumOutDir_C, "RpcEnum1", K_EnumOutDir_C, "c++ Rpc枚举的输出目录")
	flag.StringVar(&K_EnumOutDir_CS, "RpcEnum2", K_EnumOutDir_CS, "c# Rpc枚举的输出目录")
	flag.StringVar(&K_ErrOutDir_C, "Err1", K_ErrOutDir_C, "c++ 错误码的输出目录")
	flag.StringVar(&K_ErrOutDir_CS, "Err2", K_ErrOutDir_CS, "c# 错误码的输出目录")
}

func main() {
	var _regist bool
	flag.BoolVar(&_regist, "regist", false, "")
	flag.Parse() //内部获取了所有参数：os.Args[1:]

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%v: %s", r, debug.Stack())
			time.Sleep(time.Minute)
		}
	}()

	//1、收集RpcInfo -- 公共服务器、具体某游戏的业务服务器
	svrList := []string{
		"shared_svr/svr_center",
		"shared_svr/svr_chat",
		"shared_svr/svr_file",
		"shared_svr/svr_friend",
		"shared_svr/svr_gateway",
		"shared_svr/svr_gm",
		"shared_svr/svr_login",
		"shared_svr/svr_relay",
		"shared_svr/svr_save",
		"shared_svr/svr_sdk",
		"shared_svr/svr_stats",
		"shared_svr/zookeeper",
		"svr_cross",
		"svr_game",
	}
	rpcInfos := make([]*RpcInfo, len(svrList))
	for i, v := range svrList {
		rpcInfos[i] = gatherRpcInfo(v)
	}

	//2、收集Rpc函数名 -- 战斗服、客户端
	funcs1 := make(map[string]struct{}, 1024)
	for _, ptr := range rpcInfos {
		addRpc_Go(funcs1, ptr)
	}
	addRpc_C(funcs1)  //svr_battle
	addRpc_CS(funcs1) //client
	//对结果排序
	funcs := make([]string, 0, len(funcs1))
	for k := range funcs1 {
		funcs = append(funcs, k)
	}
	sort.Strings(funcs)

	//3、RpcFunc收集完毕，生成RpcEunm
	isChange := generateRpcEnum(funcs)
	if isChange {
		modules := []string{"client"}
		for _, ptr := range rpcInfos {
			modules = append(modules, ptr.Module)
		}
		//4、生成golang服务器的路由信息
		generateRpcRoute(modules, funcs)
	}

	//5、生成注册代码
	if isChange || _regist {
		for i, v := range svrList {
			generateRpcRegist(v, rpcInfos[i])
		}
	}

	//生成错误码
	generateErrCode()

	println("Generate success...")
}
