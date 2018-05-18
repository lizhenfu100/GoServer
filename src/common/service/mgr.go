package service

type ServiceObj struct {
	pObj  interface{}
	isReg bool //注册或注销对象
}
type IService interface {
	UnRegister(pObj interface{})
	Register(pObj interface{})
	RunSevice(timelapse int, timenow int64)
}

type ServiceMgr struct {
	List []IService
}

func (self *ServiceMgr) RunAllService(timelapse int, timenow int64) {
	for _, v := range self.List {
		v.RunSevice(timelapse, timenow)
	}
}
func (self *ServiceMgr) UnRegister(enum int, p interface{}) { self.List[enum].UnRegister(p) }
func (self *ServiceMgr) Register(enum int, p interface{})   { self.List[enum].Register(p) }
