package main

/*
	1、最好别用切片当参数，内部若append会丢数据
	2、range迭代器是值拷贝的，对迭代元素改写，不会影响原始数据。对比lua-pair的table引用传递
*/

import (
	"fmt"
)

type IntPair struct {
	ID  int
	Cnt int
}
type IntPairList struct {
	List []IntPair
}

func bug_Add(self []IntPair, id int, num int) int {
	if id <= 0 || num <= 0 {
		return 0
	}
	for _, v := range self {
		if v.ID == id {
			v.Cnt += num
			fmt.Println("IntPair Add =>", self) //【坑】：range产生的value，是值拷贝，改变value的数据，真正循环体self内，无变化
			return v.Cnt
		}
	}
	self = append(self, IntPair{id, num}) //【坑】：虽然切片是引用传递，但append调用后，外界实参不会改变的
	fmt.Println("IntPair New Add =>", self)
	return num
}

func (self *IntPairList) Get(id int) int {
	for _, v := range self.List {
		if v.ID == id {
			return v.Cnt
		}
	}
	return 0
}
func (self *IntPairList) Add(id, num int) int {
	if id <= 0 || num <= 0 {
		return 0
	}
	length := len(self.List)
	for i := 0; i < length; i++ {
		if self.List[i].ID == id {
			self.List[i].Cnt += num
			fmt.Println("IntPair Add =>", self.List)
			return self.List[i].Cnt
		}
	}
	self.List = append(self.List, IntPair{id, num})
	fmt.Println("IntPair New Add =>", self.List)
	return num
}
func (self *IntPairList) Del(id, num int) bool {
	if id <= 0 || num <= 0 {
		return false
	}
	length := len(self.List)
	for i := 0; i < length; i++ {
		if self.List[i].ID == id {
			if self.List[i].Cnt < num {
				return false
			} else if self.List[i].Cnt == num {
				self.List = append(self.List[:i], self.List[i+1:]...)
				fmt.Println("IntPair Delete =>", self.List)
				return true
			} else {
				self.List[i].Cnt -= num
				fmt.Println("IntPair Cut =>", self.List)
				return true
			}
		}
	}
	return false
}
