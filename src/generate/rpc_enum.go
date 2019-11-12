/***********************************************************************
* @ 生成 rpc enum
* @ brief
	1、rpc_regist.go，记录途中提取到的Rpc函数名

	2、解析 generate_out/rpc/enum/generate_rpc_enum.go，得到旧Rpc枚举表

	3、遍历Rpc函数名，同旧枚举比对，有新的才追加，保障兼容性、编译友好性

* @ 手动指定枚举顺序
	、编辑 generate_rpc_enum.go 删除尾部其它枚举，再次生成即可
	、开头几个 Rpc 是系统保留的，供底层使用，不要删除

* @ 预定义的模板全局函数
	、{{index x 1 2 3}} 返回x[1][2][3]，x必须是一个map、slice或数组

* @ author zhoumf
* @ date 2017-10-17
***********************************************************************/
package main

import (
	"common/file"
	"common/std"
	"regexp"
	"strconv"
	"strings"
)

// -------------------------------------
// 收集各处的 Rpc 函数名
func addRpc_Go(funcs map[string]struct{}, info *RpcInfo) {
	for _, v := range info.TcpRpc {
		funcs[v.Name] = struct{}{}
	}
	for _, v := range info.HttpRpc {
		funcs[v.Name] = struct{}{}
	}
	for _, v := range info.PlayerRpc {
		funcs[v.Name] = struct{}{}
	}
}
func addRpc_C(funcs map[string]struct{}) {
	reg := regexp.MustCompile(`Rpc_\w+`)
	file.ReadLine(K_RpcFuncFile_C, func(line string) {
		if ok, _ := regexp.MatchString(`Rpc_Declare\(Rpc_`, line); ok {
			funcs[reg.FindAllString(line, -1)[1]] = struct{}{}
		}
	})
}
func addRpc_CS(funcs map[string]struct{}) {
	reg := regexp.MustCompile(`Rpc_\w+`)
	file.ReadLine(K_RpcFuncFile_CS, func(line string) {
		if ok, _ := regexp.MatchString(`public void Rpc_`, line); ok {
			funcs[reg.FindAllString(line, -1)[0]] = struct{}{}
		}
	})
}
func getOldRpc() (enums []std.KeyPair, enumCnt int) {
	reg := regexp.MustCompile(`Rpc_\w+`)
	file.ReadLine(K_EnumOutDir+K_EnumFileName+".go", func(line string) {
		if ok, _ := regexp.MatchString(`Rpc_\w+\s+uint16 =`, line); ok {
			if result := reg.FindAllString(line, -1); result != nil {
				name := result[0]
				list := strings.Split(line, " ")
				rid, _ := strconv.Atoi(list[len(list)-1])
				enums, enumCnt = append(enums, std.KeyPair{name, rid}), rid+1
			}
		}
	})
	if enumCnt < 100 { //之前的预留给系统层用
		enumCnt = 100
	}
	return
}
func IsEnumIn(enums []std.KeyPair, name string) bool {
	for _, v := range enums {
		if v.Name == name {
			return true
		}
	}
	return false
}

// -------------------------------------
// 生成枚举代码
func generateRpcEnum(funcs []string) bool {
	enums, enumCnt := getOldRpc() //旧枚举，追加新的重新生成
	haveNewEnum := false
	for _, name := range funcs {
		if !IsEnumIn(enums, name) {
			enums = append(enums, std.KeyPair{name, enumCnt})
			haveNewEnum = true
			enumCnt++
		}
	}
	if !haveNewEnum { //没有新的，就不改动文件了，编译更友好
		println("no new rpc, don't change rpc_enum.h")
		return false
	}
	enums = append(enums, std.KeyPair{"RpcEnumCnt", enumCnt})

	println(K_EnumOutDir, K_EnumFileName+".go")
	file.CreateTemplate(enums, K_EnumOutDir, K_EnumFileName+".go", codeEnumTemplate_Go)

	if K_EnumOutDir_C != "" {
		println(K_EnumOutDir_C, K_EnumFileName+".h")
		file.CreateTemplate(enums, K_EnumOutDir_C, K_EnumFileName+".h", codeEnumTemplate_C)
	}
	if K_EnumOutDir_CS != "" {
		println(K_EnumOutDir_CS, K_EnumFileName+".cs")
		file.CreateTemplate(enums, K_EnumOutDir_CS, K_EnumFileName+".cs", codeEnumTemplate_CS)
	}
	return true
}

// -------------------------------------
// -- 填充模板
const (
	codeEnumTemplate_Go = `// Generated by GoServer/src/generate
// Don't edit !
package enum
const ( //the top 100 are reserved for system
	{{range $_, $v := .}}{{$v.Name}} uint16 = {{$v.ID}}
	{{end}}
)
`
	codeEnumTemplate_C = `// Generated by GoServer/src/generate
// Don't edit !
#pragma once
#define Rpc_Enum\
	{{range $_, $v := .}}_Declare({{$v.Name}}, {{$v.ID}})\
	{{end}}


#undef _Declare
#define _Declare(k, v) k = v,
enum RpcEnum:uint16 {
    Rpc_Enum
};
inline const char* RpcIdToName(int id) {
#ifdef _DEBUG
    static std::map<int, const char*> g_rpc_func;
    if (g_rpc_func.empty()) {
#undef _Declare
#define _Declare(k, v) g_rpc_func[v] = #k;
        Rpc_Enum
    }
    return g_rpc_func[id];
#else
    static char str[16];
    sprintf(str, "%d", id);
    return str;
#endif
}
`
	codeEnumTemplate_CS = `// Generated by GoServer/src/generate
// Don't edit !
public enum RpcEnum: System.UInt16 {
	{{range $_, $v := .}}{{$v.Name}} = {{$v.ID}},
	{{end}}
}
`
)
