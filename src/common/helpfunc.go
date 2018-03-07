package common

import (
	"conf"
)

type IntPair struct {
	ID  int
	Cnt int
}
type KeyPair struct {
	Name string
	ID   int
}
type Addr struct {
	IP   string
	Port uint16
}

// -------------------------------------
//! 数组封装
type (
	IntLst    []int
	UInt32Lst []uint32
)

func (self *IntLst) IsExist(v int) int {
	for i := 0; i < len(*self); i++ {
		if v == (*self)[i] {
			return i
		}
	}
	return -1
}
func (self *IntLst) Add(v int) {
	(*self) = append(*self, v)
}
func (self *IntLst) Del(i int) {
	(*self) = append((*self)[:i], (*self)[i+1:]...)
}
func (self *IntLst) Less(i, j int) bool {
	return (*self)[i] < (*self)[j]
}
func (self *IntLst) Swap(i, j int) {
	temp := (*self)[i]
	(*self)[i] = (*self)[j]
	(*self)[j] = temp
}
func (self *UInt32Lst) IsExist(v uint32) int {
	for i := 0; i < len(*self); i++ {
		if v == (*self)[i] {
			return i
		}
	}
	return -1
}
func (self *UInt32Lst) Add(v uint32) {
	(*self) = append(*self, v)
}
func (self *UInt32Lst) Del(i int) {
	(*self) = append((*self)[:i], (*self)[i+1:]...)
}
func (self *UInt32Lst) Less(i, j uint32) bool {
	return (*self)[i] < (*self)[j]
}
func (self *UInt32Lst) Swap(i, j uint32) {
	temp := (*self)[i]
	(*self)[i] = (*self)[j]
	(*self)[j] = temp
}

// -------------------------------------
//
func Assert(eq bool) {
	if conf.IsDebug && !eq {
		panic("assert")
	}
}
