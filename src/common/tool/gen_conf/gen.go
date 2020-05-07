package main

import (
	"common/file"
	"os"
)

const kTemplate = `//Generated by common/gen_conf

package {{.P}}

import "sync"

var (
	_{{.N}}		 {{.T}}
	_{{.N}}Mutex sync.Mutex
)

func {{.F}}() {{.T}} {
	_{{.N}}Mutex.Lock()
	ret := _{{.N}} //拷贝到栈上
	_{{.N}}Mutex.Unlock()
	return ret
}
func (v {{.T}}) Init() {
	_{{.N}}Mutex.Lock()
	_{{.N}} = v
	_{{.N}}Mutex.Unlock()
}
`

var g struct {
	T string //类型名
	N string //变量名
	F string //接口名
	P string //packag
}

func main() {
	g.T = os.Args[1]
	g.P = os.Args[2]
	if g.N = g.T; g.T[0] == '*' {
		g.N = g.T[1:]
	}
	if g.F = g.N; g.N[0] >= 'a' && g.N[0] <= 'z' {
		tmp := []byte(g.N)
		tmp[0] -= 'a' - 'A'
		g.F = string(tmp)
	}
	file.CreateTemplate(&g, "./", "gen_"+g.N+".go", kTemplate)
}