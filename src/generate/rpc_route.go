package main

import (
	"bytes"
	"common"
	"common/std"
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
	f.Write([]byte("func GetRpcModule(rpcEnum uint16) string {\nswitch rpcEnum {"))
	//2、case部分由模板生成
	if tpl, err := template.New(filename).Parse(funcRouteTemplate); err == nil {
		var bf bytes.Buffer
		for k, v := range splitModuleRpcs(modules, funcs) {
			if err = tpl.Execute(&bf, v); err != nil {
				panic(err.Error())
				return
			}
			buf := bf.Bytes()
			buf[len(buf)-2] = ':'
			buf[len(buf)-1] = ' '
			f.Write([]byte("\ncase"))
			f.Write(buf)
			if k == "login" {
				f.Write([]byte(`return conf.GameName`))
			} else {
				f.Write(common.S2B(fmt.Sprintf(`return "%s"`, k)))
			}
			bf.Reset()
		}
	} else {
		panic(err.Error())
		return
	}
	//3、函数收尾
	f.Write([]byte("}\nreturn \"\"}"))
}
func splitModuleRpcs(modules, funcs std.Strings) (ret map[string]std.Strings) {
	ret = make(map[string]std.Strings)
	for _, v := range funcs {
		if m := strings.Split(v, "_")[1]; modules.Index(m) >= 0 { //Rpc_后面的字符即为模块名
			if m == "nric" {
				m = "sdk" //TODO:zhoumf:待删除
			}
			vs := ret[m]
			vs.Add(v)
			ret[m] = vs
		}
	}
	return
}
