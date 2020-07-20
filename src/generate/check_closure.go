package main

import (
	"fmt"
	"strings"
)

var _spaceCnt int

//【tcp_rpc::_Handle】req、ack 不能被闭包
func CheckReq(line, fileName string, lineNum int) {
	if _spaceCnt > 0 { //处于匿名函数中
		if _spaceCnt == spaceCnt(line) {
			_spaceCnt = 0 //函数结束
		} else {
			checkReq(line, fileName, lineNum)
		}
	} else if i := strings.Index(line, "func("); i > 0 {
		_spaceCnt = spaceCnt(line) //函数开始
		checkReq(line[i:], fileName, lineNum)
	}
}
func checkReq(s, fileName string, lineNum int) {
	if strings.Index(s, "req") >= 0 {
		fmt.Println("Error: upvalue req:", fileName, lineNum)
	}
}
func spaceCnt(line string) (ret int) { //本行起始，有多少空格
	for _, v := range line {
		if v == '	' {
			ret += 1
		} else {
			return
		}
	}
	return
}
