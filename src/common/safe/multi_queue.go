/***********************************************************************
* @ 多重队列，交换缓冲区，优化数据竞争
* @ brief
	1、适用于生产者-消费者模型，生产者一般都是多个，消费者默认单个
		· 消费者也可多个，比如工作线程池……但须额外处理竞态
		· 多消费者，须扩增multi数目

	2、准备多个队列，选取一个作为写者，供生产者写入数据；此时仅生产者之间的竞态
	3、消费者取数据时，加锁，将原来的写者取出，并且新从multi中选个队列，作新的写者

	4、【多消费者时，可能消费数据过快，muti分配队列错乱，各消费者/生产者拿到的并不是单owner队列】
		· 所以单消费者更健壮
		· 可消费者取出后，再由工作线程池从消费者队列中取，转移竞态

* @ 参考
	https://github.com/lni/dragonboat/blob/master/queue.go
	https://github.com/3workman/CXServer/blob/master/src/common/Log/AsyncLog.h

* @ author zhoumf
* @ date 2019-3-27
***********************************************************************/
package safe

import "sync"

// ------------------------------------------------------------
type MultiQueue struct {
	sync.Mutex
	multi       [2][]interface{} //须比消费者数目多1，生产者共用一个写
	writer      []interface{}    //引用自multi，生产者们往里写入
	writerCycle uint8            //multi中循环挑选，作为writer
	stop        bool
	kSize       uint32
	wpos        uint32
}

func (self *MultiQueue) Init(size uint32) {
	self.kSize = size
	for i := 0; i < len(self.multi); i++ {
		self.multi[i] = make([]interface{}, size)
	}
	self.writer = self.multi[0]
}
func (self *MultiQueue) Close() {
	self.Lock()
	self.stop = true
	self.Unlock()
}
func (self *MultiQueue) PendingPos() (ret uint32) {
	self.Lock()
	ret = self.wpos
	self.Unlock()
	return
}
func (self *MultiQueue) Put(val interface{}) (bool, bool) {
	self.Lock()
	if self.wpos >= self.kSize {
		self.Unlock()
		return false, self.stop
	}
	if self.stop {
		self.Unlock()
		return false, true
	}
	self.writer[self.wpos] = val
	self.wpos++
	self.Unlock()
	return true, false
}
func (self *MultiQueue) Get() (ret []interface{}) {
	self.Lock()
	ret = self.writer[:self.wpos]
	self.writerCycle = (self.writerCycle + 1) % uint8(len(self.multi))
	self.writer = self.multi[self.writerCycle] //change writer
	self.wpos = 0
	self.Unlock()
	return
}

// ------------------------------------------------------------
type MultiQueueEx struct {
	sync.Mutex
	multi          [2][]interface{} //须比消费者数目多1，生产者共用一个写
	writer         []interface{}    //引用自multi，生产者们往里写入
	writerCycle    uint8            //multi中循环挑选，作为writer
	stop           bool
	pause          bool
	kLazyFreeCycle uint8 //惰性清理引用，让gc回收；队列极长时，可能尾部数据一直被引用着，无法gc
	freeCycle      uint8
	kSize          uint32
	wpos           uint32
	wposOld        uint32 //上次写了多少，用于减少清理数目
}

func (self *MultiQueueEx) Init(size uint32, lazyFreeCycle uint8) {
	self.kSize = size
	self.kLazyFreeCycle = lazyFreeCycle
	for i := 0; i < len(self.multi); i++ {
		self.multi[i] = make([]interface{}, size)
	}
	self.writer = self.multi[0]
}
func (self *MultiQueueEx) Close() {
	self.Lock()
	self.stop = true
	self.Unlock()
}
func (self *MultiQueueEx) Put(val interface{}) (bool, bool) {
	self.Lock()
	if self.pause || self.wpos >= self.kSize {
		self.Unlock()
		return false, self.stop
	}
	if self.stop {
		self.Unlock()
		return false, true
	}
	self.writer[self.wpos] = val
	self.wpos++
	self.Unlock()
	return true, false
}
func (self *MultiQueueEx) Get(pause bool) (ret []interface{}) {
	self.Lock()
	self.pause = pause
	ret = self.writer[:self.wpos]
	self.writerCycle = (self.writerCycle + 1) % uint8(len(self.multi))
	self.writer = self.multi[self.writerCycle] //change writer
	self.free()
	self.wposOld, self.wpos = self.wpos, 0
	self.Unlock()
	return
}
func (self *MultiQueueEx) free() {
	if self.kLazyFreeCycle > 0 {
		self.freeCycle++
		oldq := self.writer
		if self.kLazyFreeCycle == 1 {
			for i := uint32(0); i < self.wposOld; i++ {
				oldq[i] = nil
			}
		} else if self.freeCycle%self.kLazyFreeCycle == 0 {
			for i := uint32(0); i < self.kSize; i++ {
				oldq[i] = nil
			}
		}
	}
}

// ------------------------------------------------------------
type TReadyIDs struct {
	sync.Mutex
	multi      [2]map[uint32]Empty //须比消费者数目多1，生产者共用一个写入
	ready      map[uint32]Empty
	readyCycle uint8
}
type Empty struct{}

func (self *TReadyIDs) Init() {
	for i := 0; i < len(self.multi); i++ {
		self.multi[i] = make(map[uint32]Empty)
	}
	self.ready = self.multi[0]
}
func (self *TReadyIDs) SetReady(clusterID uint32) {
	self.Lock()
	self.ready[clusterID] = Empty{}
	self.Unlock()
}
func (self *TReadyIDs) GetReady() (ret map[uint32]Empty) {
	self.Lock()
	ret = self.ready
	self.readyCycle = (self.readyCycle + 1) % uint8(len(self.multi))
	self.ready = self.multi[self.readyCycle] //change ready
	for k := range self.ready {
		delete(self.ready, k)
	}
	self.Unlock()
	return
}
