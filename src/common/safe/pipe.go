package safe

import "sync"

type Pipe struct { //非阻塞
	sync.Mutex
	multi       [2][]interface{}
	writer      []interface{}
	writerCycle uint8
}
type Chan struct { //阻塞
	sync.Cond
	Pipe
}

func (self *Pipe) Init(cap uint32) {
	for i := 0; i < len(self.multi); i++ {
		self.multi[i] = make([]interface{}, 0, cap)
	}
	self.writer = self.multi[0]
}
func (self *Pipe) Add(v interface{}) {
	self.Lock()
	self.writer = append(self.writer, v)
	self.Unlock()
}
func (self *Pipe) Get() (ret []interface{}) {
	self.Lock()
	ret = self._get()
	self.Unlock()
	return
}
func (self *Pipe) _get() (ret []interface{}) {
	ret = self.writer
	self.writerCycle = (self.writerCycle + 1) % uint8(len(self.multi))
	self.writer = self.multi[self.writerCycle] //change writer
	self.writer = self.writer[:0]
	return ret
}
func (self *Chan) Init(cap uint32) {
	self.Pipe.Init(cap)
	self.Cond.L = &self.Mutex
}
func (self *Chan) Add(v interface{}) {
	self.Lock()
	self.writer = append(self.writer, v)
	self.Unlock()
	self.Signal()
}
func (self *Chan) Get() (ret []interface{}) {
	self.Lock()
	for len(self.writer) == 0 {
		self.Wait()
	}
	ret = self._get()
	self.Unlock()
	return
}
