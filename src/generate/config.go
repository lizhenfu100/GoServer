package main

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
