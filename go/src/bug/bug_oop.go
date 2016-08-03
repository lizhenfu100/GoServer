package bug

import (
	"fmt"
)

type FunOOP interface {
	Name() string
	PrintName()
}
type TParent struct {
}
type TSubclass struct {
	TParent
}

func (self *TParent) Name() string {
	return "a"
}
func (self *TParent) PrintName() { // OOP中，传入的self实际是指向子类的，所以self.Name()调用的亦是子类方法
	fmt.Println(self.Name()) // 但golang是纯粹的对象组合，这里调用的仍是父类方法
}
func (self *TSubclass) Name() string {
	return "b"
}

/*
	1、golang中运用的是组合方式，“面向对象”实际是匿名组合
	2、函数调用时，若本类未实现该函数，会寻找匿名对象中的同名函数，实际也是调用的匿名组合对象的Function
*/
func test_OOP() {
	// 输出的是a，违反了通常OOP中会输出b的习惯
	var obj FunOOP = &TSubclass{}
	obj.PrintName() // 实际为：obj."".PrintName()  访问匿名TParent对象再调用的
}
