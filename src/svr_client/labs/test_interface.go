package main

import (
	"fmt"
)

// interface只关心接口“输入什么、输出什么”，不关心接口具体属于谁，运行时能调到即可
// 基于组合：因没有继承体系，一个对象无法访问“父对象”的数据，只能通过包含(组件)将所需结构加入自己内部
/* 继承体系，相比于Go的组合，有什么要吐槽的？
1、继承的强绑定体现在哪？
	子类可自由访问父类数据，编码上无法察觉，重构/维护困难高（一眼看过去鬼知道this._data到底哪个继承块的）
   	由于这个原因，父类数据可能散在多个子类中到处用……怎么改~
2、组合的话，多了一层，要访问“父类”(此时该称组件)数据，必须通过this.parent._data，所有访问都有同一个入口
*/

type FunInterface interface {
	Fun1()
	Fun2(int) bool
}
type T1 struct {
	a int
	b string
}
type T2 struct {
	c string
}

func (self *T1) Fun1() {
	fmt.Printf("aaaaa")
}
func (self *T1) Fun2(a int) bool {
	fmt.Printf("aaaaa22222")
	return true
}
func (self *T2) Fun1() {
	fmt.Printf("bbbbb")
}
func (self *T2) Fun2(a int) bool {
	fmt.Printf("bbbbb22222")
	return true
}

func TestInterface() {
	t1 := T1{}
	t2 := T2{}

	var fun FunInterface = &t2
	fun.Fun2(1)

	fun = &t1
	fun.Fun1()
}

// ------------------------------------------------------------
// 接口/类型查询
func TestInterfaceSelect() {
	var i interface{} = T1{1, "龙蛋"}

	switch v := i.(type) {
	case int32:
		fmt.Println("int32")
	case FunInterface:
		fmt.Println("FunInterface")
	case T1:
		fmt.Println(i.(T1).b, v.b)
	case interface{}:
		fmt.Println("interface{}", v) // 这条也能匹配上
	}

	//! 直接判断interface是否为某类型
	if v, ok := i.(T2); ok {
		fmt.Println(ok, v.c)
	} else {
		fmt.Println(ok, v) // false {}
	}
	if v, ok := i.(FunInterface); ok {
		fmt.Println(ok, v.Fun2(11))
	} else {
		fmt.Println(ok, v) // false <nil>
	}
}
