package safe

import (
	"sync"
)

// 放开容量限制的channel，添加不阻塞，接收可能阻塞等待
type Pipe struct {
	sync.Mutex
	cond sync.Cond
	list []interface{}
}

func NewPipe() *Pipe {
	self := new(Pipe)
	self.cond.L = &self.Mutex
	return self
}

// 添加时不会发生阻塞
func (self *Pipe) Add(msg interface{}) {
	self.Lock()
	self.list = append(self.list, msg)
	self.Unlock()
	self.cond.Signal()
}

// 如果没有数据，发生阻塞
func (self *Pipe) Pick(retList *[]interface{}) (exit bool) {
	self.Lock()
	for len(self.list) == 0 {
		self.cond.Wait()
	}
	for _, v := range self.list { //复制出队列
		if v == nil {
			exit = true
			break
		} else {
			*retList = append(*retList, v)
		}
	}
	self.list = self.list[:0] //清空
	self.Unlock()
	return
}
func (self *Pipe) Clear() { self.list = self.list[:0] }
