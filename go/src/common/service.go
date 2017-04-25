package common

import (
	"time"
)

type TChanObj struct {
	pObj  interface{}
	isReg bool
}
type ServicePatch struct {
	callback  func(interface{})
	timeWait  int // msec
	kTimeAll  int // msec
	runPos    int
	obj_lst   []interface{}
	writeChan chan TChanObj
}

func NewServicePatch(fun func(interface{}), timeAllMsec int) *ServicePatch {
	ptr := new(ServicePatch)
	ptr.callback = fun
	ptr.kTimeAll = timeAllMsec
	ptr.writeChan = make(chan TChanObj, 1024)
	go ptr._loop()
	return ptr
}
func (self *ServicePatch) UnRegister(pObj interface{}) { self.writeChan <- TChanObj{pObj, false} }
func (self *ServicePatch) Register(pObj interface{})   { self.writeChan <- TChanObj{pObj, true} }
func (self *ServicePatch) RunSevice()                  { self.writeChan <- TChanObj{nil, false} }
func (self *ServicePatch) _loop() {
	t1 := time.Now()
	for data := range self.writeChan {
		if data.pObj != nil {
			if data.isReg {
				self._doRegister(data.pObj)
			} else {
				self._doUnRegister(data.pObj)
			}
		}
		self._runSevice(time.Now().Sub(t1).Nanoseconds() / int64(time.Millisecond))
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
func (self *ServicePatch) _runSevice(timelapse int64) {
	totalCnt := len(self.obj_lst)
	if totalCnt <= 0 {
		return
	}
	//! 单位时长里要处理的个数，可能大于列表中obj总数，比如服务器卡顿很久，得追帧
	self.timeWait += int(timelapse)
	runCnt := self.timeWait * totalCnt / self.kTimeAll
	if runCnt == 0 {
		return
	}
	//! 处理一个的时长
	temp := self.kTimeAll / totalCnt
	//! 更新等待时间(须小于"处理一个的时长")：对"处理一个的时长"取模(除法的非零保护)
	if temp > 0 {
		self.timeWait %= temp
	} else {
		self.timeWait = 0
	}

	for i := 0; i < runCnt; i++ {
		ptr := self.obj_lst[self.runPos]
		self.runPos++
		if self.runPos == totalCnt { //到末尾了，回到队头
			self.runPos = 0
		}
		self.callback(ptr)
	}
}
