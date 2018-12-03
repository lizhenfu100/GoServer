package service

// -------------------------------------
// -- 周期性调用传入的函数
type TimePair struct {
	exeTime int64
	obj     interface{}
}
type ServiceVec struct {
	cb     func(interface{})
	runPos int
	cdMs   int
	objs   []TimePair
}

func NewServiceVec(fun func(interface{}), cdMs int) *ServiceVec {
	ptr := new(ServiceVec)
	ptr.cdMs = cdMs
	ptr.cb = fun
	return ptr
}
func (self *ServiceVec) Register(pObj interface{}) {
	self.objs = append(self.objs, TimePair{0, pObj})
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
			runObj := it.obj
			if self.runPos++; self.runPos >= len(self.objs) { //到末尾了，回到队头
				self.runPos = 0
			}
			//timeDiff := int(timenow-it.exeTime) + self.cdMs //实际经过的时间间隔
			it.exeTime = timenow + int64(self.cdMs)
			self.cb(runObj) //里头可能把自己删掉，runPos指向改变，it可能失效
		} else {
			break
		}
	}
}
