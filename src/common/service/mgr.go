package service

import "common/safe"

type IService interface {
	UnRegister(pObj interface{})
	Register(pObj interface{})
	RunSevice(timelapse int, timenow int64)
}
type obj struct {
	ptr   interface{}
	enum  int
	isReg bool //注册或注销对象
}
type ServiceMgr struct {
	list  []IService
	queue safe.SafeQueue
}

func (self *ServiceMgr) Init(cap uint32, list []IService) {
	self.queue.Init(cap)
	self.list = list
}
func (self *ServiceMgr) Cap() uint32 { return self.queue.Cap() }

func (self *ServiceMgr) RunAllService(timelapse int, timenow int64) {
	for {
		if v, ok, _ := self.queue.Get(); ok {
			if v := v.(obj); v.isReg {
				self.list[v.enum].Register(v.ptr)
			} else {
				self.list[v.enum].UnRegister(v.ptr)
			}
		} else {
			break
		}
	}
	for _, v := range self.list {
		v.RunSevice(timelapse, timenow)
	}
}
func (self *ServiceMgr) UnRegister(enum int, p interface{}) bool {
	ok, _ := self.queue.Put(obj{p, enum, false})
	return ok
}
func (self *ServiceMgr) Register(enum int, p interface{}) bool { //【防止多次注册】
	ok, _ := self.queue.Put(obj{p, enum, true})
	return ok
}
