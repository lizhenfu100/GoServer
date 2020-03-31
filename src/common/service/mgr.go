package service

import (
	"common/safe"
)

type IService interface {
	UnRegister(p interface{})
	Register(p interface{})
	RunSevice(timelapse int, timenow int64)
}
type ServiceMgr struct {
	safe.Pipe
	list []IService
}

func (self *ServiceMgr) Init(cap uint32, list []IService) {
	self.Pipe.Init(cap)
	self.list = list
}
func (self *ServiceMgr) RunAllService(timelapse int, timenow int64) {
	for _, v := range self.Get() {
		if v := v.(obj); v.isReg {
			self.list[v.enum].Register(v.ptr)
		} else {
			self.list[v.enum].UnRegister(v.ptr)
		}
	}
	for _, v := range self.list {
		v.RunSevice(timelapse, timenow)
	}
}

//【防止多次注册】
func (self *ServiceMgr) Register(enum int, p interface{})   { self.Add(obj{p, enum, true}) }
func (self *ServiceMgr) UnRegister(enum int, p interface{}) { self.Add(obj{p, enum, false}) }

type obj struct {
	ptr   interface{}
	enum  int
	isReg bool
}
