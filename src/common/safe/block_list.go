package safe

import (
	"common"
	"sync"
)

type BlockList struct {
	sync.Mutex
	cond *sync.Cond
	list []*common.NetPack
}

func NewBlockList() *BlockList {
	self := new(BlockList)
	self.cond = sync.NewCond(&self.Mutex)
	return self
}
func (self *BlockList) Add(p *common.NetPack) {
	self.Lock()
	self.list = append(self.list, p)
	self.Unlock()
	self.cond.Signal()
}
func (self *BlockList) Del(p *common.NetPack) {
	self.Lock()
	defer self.Unlock()
	for i, v := range self.list {
		if v == p {
			self.list = append(self.list[:i], self.list[i+1:]...)
			return
		}
	}
}
func (self *BlockList) MoveToStack(list *[]*common.NetPack) {
	self.Lock()
	defer self.Unlock()
	for len(self.list) == 0 {
		self.cond.Wait()
	}
	// copy on write
	// c++中可用shared_ptr.unique()判断当前是否仅一个操作者
	// 若是则不必拷贝，直接加锁处理
	*list = append(*list, self.list...)
	self.list = self.list[:0]
}
