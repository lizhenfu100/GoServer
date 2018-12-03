package service

// -------------------------------------
// -- 花一段时长，遍历完所有对象
type ServicePatch struct {
	cb       func(interface{})
	timeWait int // msec
	kTimeAll int // msec
	runPos   int
	objs     []interface{}
}

func NewServicePatch(fun func(interface{}), timeAllMsec int) *ServicePatch {
	ptr := new(ServicePatch)
	ptr.cb = fun
	ptr.kTimeAll = timeAllMsec
	return ptr
}
func (self *ServicePatch) Register(pObj interface{}) { self.objs = append(self.objs, pObj) }
func (self *ServicePatch) UnRegister(pObj interface{}) {
	for i, ptr := range self.objs {
		if ptr == pObj {
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
func (self *ServicePatch) RunSevice(timelapse int, timenow int64) {
	if len(self.objs) <= 0 {
		return
	}
	//! 单位时长里要处理的个数，可能大于列表中obj总数，比如服务器卡顿很久，得追帧
	self.timeWait += timelapse
	runCnt := self.timeWait * len(self.objs) / self.kTimeAll
	if runCnt == 0 {
		return
	}
	//! 处理一个的时长
	temp := self.kTimeAll / len(self.objs)
	//! 更新等待时间(须小于"处理一个的时长")：对"处理一个的时长"取模(除法的非零保护)
	if temp > 0 {
		self.timeWait %= temp
	} else {
		self.timeWait = 0
	}

	for i := 0; i < runCnt; i++ {
		obj := self.objs[self.runPos]
		if self.runPos++; self.runPos >= len(self.objs) { //到末尾了，回到队头
			self.runPos = 0
		}
		self.cb(obj)
	}
}
