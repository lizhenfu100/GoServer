package main

import (
	"fmt"
	"testing"
	"unsafe"
)

//go test -test.bench="."
var g_x []byte
var g_xx []byte

func Init() {
	g_x = []byte("aa.bb.*")
	g_xx = []byte("aa.bb.**")
}

func Format1() string {
	idx := byte(2)
	str := fmt.Sprintf("aa.bb.%d", idx)
	return str
}
func Format2() string { //快了20倍！
	idx := byte(23)
	if idx < 10 {
		// x := []byte("aa.bb.*") //优化：换成静态生存期
		x := g_x
		x[len(x)-1] = idx + '0'
		str := *(*string)(unsafe.Pointer(&x))
		return str
	} else if idx < 100 {
		// x := []byte("aa.bb.**")
		x := g_xx
		x[len(x)-1] = idx%10 + '0'
		x[len(x)-2] = idx/10 + '0'
		str := *(*string)(unsafe.Pointer(&x))
		return str
	} else {
		panic("idx too big!")
		return ""
	}
}

func TestString1(t *testing.T) {
	fmt.Println(Format1())
}
func TestString2(t *testing.T) {
	Init()
	fmt.Println(Format2())
}
func Benchmark_Str1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Format1()
	}
}
func Benchmark_Str2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Format2()
	}
}
