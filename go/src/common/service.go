package common

type ServiceObj struct {
	pObj  interface{}
	isReg bool
}

// -------------------------------------
// -- 花一段时长，遍历完所有对象
type ServicePatch struct {
	callback  func(interface{})
	timeWait  int // msec
	kTimeAll  int // msec
	runPos    int
	obj_lst   []interface{}
	writeChan chan ServiceObj
}

func NewServicePatch(fun func(interface{}), timeAllMsec int) *ServicePatch {
	ptr := new(ServicePatch)
	ptr.callback = fun
	ptr.kTimeAll = timeAllMsec
	ptr.writeChan = make(chan ServiceObj, 64)
	return ptr
}
func (self *ServicePatch) UnRegister(pObj interface{}) { self.writeChan <- ServiceObj{pObj, false} }
func (self *ServicePatch) Register(pObj interface{})   { self.writeChan <- ServiceObj{pObj, true} }
func (self *ServicePatch) RunSevice(timelapse int) {
	for {
		select {
		case data := <-self.writeChan:
			if data.isReg {
				self._doRegister(data.pObj)
			} else {
				self._doUnRegister(data.pObj)
			}
		default:
			self._runSevice(timelapse)
			return
		}
	}
}
func (self *ServicePatch) _doRegister(pObj interface{}) { self.obj_lst = append(self.obj_lst, pObj) }
func (self *ServicePatch) _doUnRegister(pObj interface{}) {
	for i, ptr := range self.obj_lst {
		if ptr == pObj {
			self.obj_lst = append(self.obj_lst[:i], self.obj_lst[i+1:]...)
			if i < self.runPos {
				self.runPos--
			}
			return
		}
	}
}
func (self *ServicePatch) _runSevice(timelapse int) {
	// Notice: 中途可能增删，长度会变，每次算len安全
	// totalCnt := len(self.obj_lst)
	// if totalCnt <= 0 {
	// 	return
	// }
	//! 单位时长里要处理的个数，可能大于列表中obj总数，比如服务器卡顿很久，得追帧
	self.timeWait += timelapse
	runCnt := self.timeWait * len(self.obj_lst) / self.kTimeAll
	if runCnt == 0 {
		return
	}
	//! 处理一个的时长
	temp := self.kTimeAll / len(self.obj_lst)
	//! 更新等待时间(须小于"处理一个的时长")：对"处理一个的时长"取模(除法的非零保护)
	if temp > 0 {
		self.timeWait %= temp
	} else {
		self.timeWait = 0
	}

	for i := 0; i < runCnt; i++ {
		ptr := self.obj_lst[self.runPos]
		self.runPos++
		if self.runPos == len(self.obj_lst) { //到末尾了，回到队头
			self.runPos = 0
		}
		self.callback(ptr)
	}
}

// -------------------------------------
// --
type TimePair struct {
	time int64
	ptr  interface{}
}
type ServiceList struct {
	callback  func(interface{})
	runPos    int
	cdMs      int
	obj_lst   []TimePair
	writeChan chan ServiceObj
}

func NewServiceList(fun func(interface{}), cdMs int) *ServiceList {
	ptr := new(ServiceList)
	ptr.cdMs = cdMs
	ptr.callback = fun
	ptr.writeChan = make(chan ServiceObj, 64)
	return ptr
}
func (self *ServiceList) UnRegister(pObj interface{}) { self.writeChan <- ServiceObj{pObj, false} }
func (self *ServiceList) Register(pObj interface{})   { self.writeChan <- ServiceObj{pObj, true} }
func (self *ServiceList) RunSevice(timenow int64) {
	for {
		select {
		case data := <-self.writeChan:
			if data.isReg {
				self._doRegister(data.pObj)
			} else {
				self._doUnRegister(data.pObj)
			}
		default:
			self._runSevice(timenow)
			return
		}
	}
}
func (self *ServiceList) _doRegister(pObj interface{}) {
	self.obj_lst = append(self.obj_lst, TimePair{0, pObj})
}
func (self *ServiceList) _doUnRegister(pObj interface{}) {
	for i, it := range self.obj_lst {
		if it.ptr == pObj {
			self.obj_lst = append(self.obj_lst[:i], self.obj_lst[i+1:]...)
			if i < self.runPos {
				self.runPos--
			}
			return
		}
	}
}
func (self *ServiceList) _runSevice(timenow int64) {
	for self.runPos != len(self.obj_lst) {
		it := &self.obj_lst[self.runPos]
		if it.time <= timenow {
			runPtr := it.ptr
			self.callback(it.ptr) //里头可能删掉自己，it指向改变
			if it.ptr == runPtr {
				it.time = timenow + int64(self.cdMs)
			}
			self.runPos++
			if self.runPos == len(self.obj_lst) { //到末尾了，回到队头
				self.runPos = 0
			}
		} else {
			break
		}
	}
}
