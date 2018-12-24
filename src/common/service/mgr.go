package service

type Obj struct {
	ptr   interface{}
	enum  int
	isReg bool //注册或注销对象
}
type IService interface {
	UnRegister(pObj interface{})
	Register(pObj interface{})
	RunSevice(timelapse int, timenow int64)
}
type ServiceMgr struct {
	Chan chan Obj
	List []IService
}

func (self *ServiceMgr) RunAllService(timelapse int, timenow int64) {
	for {
		select {
		case data := <-self.Chan:
			if data.isReg {
				self.List[data.enum].Register(data.ptr)
			} else {
				self.List[data.enum].UnRegister(data.ptr)
			}
		default:
			for _, v := range self.List {
				v.RunSevice(timelapse, timenow)
			}
			return
		}
	}
}
func (self *ServiceMgr) UnRegister(enum int, p interface{}) { self.Chan <- Obj{p, enum, false} }
func (self *ServiceMgr) Register(enum int, p interface{})   { self.Chan <- Obj{p, enum, true} }
