package common

import (
	"sync"
)

type BlockList struct {
	list  []*NetPack
	mutex sync.Mutex
	cond  *sync.Cond
}

func NewBlockList() *BlockList {
	self := new(BlockList)
	self.cond = sync.NewCond(&self.mutex)
	return self
}
func (self *BlockList) Add(p *NetPack) {
	self.mutex.Lock()
	self.list = append(self.list, p)
	self.mutex.Unlock()
	self.cond.Signal()
}
func (self *BlockList) Del(p *NetPack) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	for i, v := range self.list {
		if v == p {
			self.list = append(self.list[:i], self.list[i+1:]...)
			return
		}
	}
}
func (self *BlockList) MoveToStack(list *[]*NetPack) {
	self.mutex.Lock()
	for len(self.list) == 0 {
		self.cond.Wait()
	}
	// copy on write
	// c++中可用shared_ptr.unique()判断当前是否仅一个操作者
	// 若是则不必拷贝，直接加锁处理
	// 不过拷贝到栈上后，就不必加锁了
	*list = append(*list, self.list...)
	self.list = self.list[:0]
	self.mutex.Unlock()
}
