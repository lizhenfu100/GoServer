package main

import (
	"bytes"
	"common"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// -------------------------------------
// -- 生成RpcEnum同模块的对应关系，供gateway路由
const funcRouteTemplate = `
{{range $_, $v := .}}	{{$v}},
{{end}}`

func generateRpcRoute(modules, funcs []string) {
	filename := K_EnumFileName + ".go"
	f, err := os.OpenFile(K_EnumOutDir+filename, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
		return
	}
	defer f.Close()

	//1、先写入函数的switch部分
	f.Write(common.S2B("func GetRpcModule(rpcEnum uint16) string {\nswitch rpcEnum {"))

	//2、case部分由模板生成
	if tpl, err := template.New(filename).Parse(funcRouteTemplate); err == nil {
		var bf bytes.Buffer
		for _, module := range modules {
			if list := getModuleRpcs(module, funcs); len(list) > 0 {
				if err = tpl.Execute(&bf, list); err != nil {
					panic(err.Error())
					return
				}
				buf := bf.Bytes()
				buf[len(buf)-2] = ':'
				buf[len(buf)-1] = ' '
				f.Write(common.S2B("\ncase "))
				f.Write(buf)
				f.Write(common.S2B(fmt.Sprintf(`return "%s"`, module)))
				bf.Reset()
			}
		}
	} else {
		panic(err.Error())
		return
	}

	//3、函数收尾
	f.Write(common.S2B("}\nreturn \"\"}"))
}
func getModuleRpcs(module string, funcs []string) (ret []string) {
	for _, name := range funcs {
		if strings.Split(name, "_")[1] == module { //Rpc_后面的字符即为模块名
			ret = append(ret, name)
		}
	}
	return
}
