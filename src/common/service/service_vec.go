package service

// -------------------------------------
// -- 固定周期回调
type ServiceVec struct {
	cb       func(interface{})
	interval int //每隔几毫秒执行回调
	runPos   int
	objs     []TimePair
}
type TimePair struct {
	obj     interface{}
	exeTime int64 //obj待回调的时刻
}

func NewServiceVec(fun func(interface{}), interval int) *ServiceVec {
	ptr := new(ServiceVec)
	ptr.interval = interval
	ptr.cb = fun
	return ptr
}
func (self *ServiceVec) Register(pObj interface{}) {
	self.objs = append(self.objs, TimePair{pObj, 0})
}
func (self *ServiceVec) UnRegister(pObj interface{}) {
	for i, it := range self.objs {
		if it.obj == pObj {
			self.objs = append(self.objs[:i], self.objs[i+1:]...)
			if i < self.runPos {
				self.runPos--
			} else if self.runPos >= len(self.objs) {
				self.runPos = 0
			}
			return
		}
	}
}
func (self *ServiceVec) RunSevice(timelapse int, timenow int64) {
	for self.runPos < len(self.objs) {
		if it := &self.objs[self.runPos]; it.exeTime <= timenow { //callback后it可能失效
			if self.runPos++; self.runPos >= len(self.objs) { //到末尾了，回到队头
				self.runPos = 0
			}
			//timeDiff := int(timenow-it.exeTime) + self.interval //实际经过的间隔
			it.exeTime = timenow + int64(self.interval)
			self.cb(it.obj) //里头可能把自己删掉，runPos指向改变，it可能失效
		} else {
			break
		}
	}
}
