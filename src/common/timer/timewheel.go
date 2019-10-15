/***********************************************************************
* @ 【单线程】时间轮
* @ brief

* @ go 原生定时器    https://www.jianshu.com/p/ac94989215d6
	、time.After、time.Tick
		· 不会粗暴的增加协程，都是打包成runtimeTimer交给底层执行
	、time.AfterFunc
		· 为确保外界传入的回调不会阻塞内部定时器协程，time.AfterFunc回调都是另开协程执行的
		· time.AfterFunc -> goFunc

* @ author zhoumf
* @ date 2017-5-11
***********************************************************************/
package timer

import (
	"common/assert"
	"common/safe"
	"gamelog"
	"time"
)

const (
	kWheelNum    = len(kWheelBit)
	kTimeTickLen = 25 //一格的刻度 ms
)

var (
	kWheelBit  = [...]uint{8, 6, 6, 6, 5} //每一级的槽位数量；用了累计位移，总和不可超32
	kWheelSize [kWheelNum]uint
	kWheelCap  [kWheelNum]uint
	G_TimerMgr SafeMgr
)

func init() {
	G_TimerMgr.Init(10240)
}

// ------------------------------------------------------------
// node
type TimeNode struct {
	prev     *TimeNode
	next     *TimeNode
	timeDead int64
	interval int
	total    int
	callback func()
}

func (self *TimeNode) init() {
	self.prev = self
	self.next = self //circle
}
func newNode(cb func(), delay, cd, total float32) *TimeNode {
	p := &TimeNode{
		timeDead: time.Now().UnixNano()/int64(time.Millisecond) + int64(delay*1000),
		interval: int(cd * 1000),
		total:    int(total * 1000),
		callback: cb,
	}
	p.init()
	return p
}
func (self *TimeNode) _Callback() {
	isJoin := false
	if self.total < 0 {
		isJoin = true
	} else if self.total -= self.interval; self.total > 0 {
		isJoin = true
	}
	if isJoin {
		self.timeDead += int64(self.interval)
		G_TimerMgr.timeWheel._AddTimerNode(self.interval, self)
	}
	self.callback()
}
func (self *TimeNode) Stop() bool { return G_TimerMgr.DelTimer(self) }

// ------------------------------------------------------------
// wheel
type stWheel struct {
	//每个slot维护的node链表为一个环，如此可以简化插入删除的操作
	//slot.next为node链表中第一个节点，prev为node的最后一个节点
	slots   []TimeNode
	slotIdx uint
}

func (self *stWheel) init(size uint) {
	self.slots = make([]TimeNode, size)
	for i := uint(0); i < size; i++ {
		self.slots[i].init()
	}
}
func (self *stWheel) GetCurSlot() *TimeNode {
	return &self.slots[self.slotIdx]
}
func (self *stWheel) size() uint { return uint(len(self.slots)) }

// ------------------------------------------------------------
type timeWheel struct {
	wheels     [kWheelNum]stWheel
	readyNode  TimeNode
	timeElapse int
}

func (self *timeWheel) Init() {
	self.readyNode.init()
	for i := 0; i < kWheelNum; i++ {
		var wheelCap uint
		for j := 0; j <= i; j++ {
			wheelCap += kWheelBit[j]
		}
		assert.True(wheelCap < 32)
		kWheelCap[i] = 1 << wheelCap
		kWheelSize[i] = 1 << kWheelBit[i]
		self.wheels[i].init(kWheelSize[i])
	}
}
func (self *timeWheel) Refresh(timeElapse int, timenow int64) {
	self.timeElapse += timeElapse
	tickCnt := self.timeElapse / kTimeTickLen
	self.timeElapse %= kTimeTickLen
	for i := 0; i < tickCnt; i++ { //扫过的slot均超时
		isCascade := false
		wheel := &self.wheels[0]
		slot := wheel.GetCurSlot()
		if wheel.slotIdx++; wheel.slotIdx >= wheel.size() {
			wheel.slotIdx = 0
			isCascade = true
		}
		node := slot.next
		slot.next, slot.prev = slot, slot //清空当前格子
		for node != slot {                //环形链表遍历
			tmp := node
			node = node.next //得放在前面，后续函数调用，可能会更改node的链接关系
			self._AddToReadyNode(tmp)
		}
		if isCascade {
			self._Cascade(1, timenow) //跳级
		}
	}
	self._DoTimeOutCallBack()
}

// total：负值表示无限循环
func (self *timeWheel) AddTimerSec(cb func(), delay, cd, total float32) *TimeNode {
	p := newNode(cb, delay, cd, total)
	self._AddTimerNode(int(delay*1000), p)
	return p
}
func (self *timeWheel) _AddTimerNode(msec int, node *TimeNode) {
	var slot *TimeNode
	tickCnt := uint(msec / kTimeTickLen)
	if tickCnt < kWheelCap[0] {
		idx := (self.wheels[0].slotIdx + tickCnt) & (kWheelSize[0] - 1) //2的N次幂位操作取余
		slot = &self.wheels[0].slots[idx]
	} else {
		for i := 1; i < kWheelNum; i++ {
			if tickCnt < kWheelCap[i] {
				preCap := kWheelCap[i-1] //上一级总容量即为本级的一格容量
				idx := (self.wheels[i].slotIdx + tickCnt/preCap - 1) & (kWheelSize[i] - 1)
				slot = &self.wheels[i].slots[idx]
				break
			}
		}
	}
	node.prev = slot.prev //插入格子的prev位置(尾节点)
	node.prev.next = node
	node.next = slot
	slot.prev = node
}
func (self *timeWheel) DelTimer(node *TimeNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
	node.prev, node.next = node, node //circle
}
func (self *timeWheel) _Cascade(wheelIdx int, timenow int64) {
	if wheelIdx < 1 || wheelIdx >= kWheelNum {
		return
	}
	isCascade := false
	wheel := &self.wheels[wheelIdx]
	slot := wheel.GetCurSlot()
	//【Bug】须先更新槽位————扫格子时加新Node，不能再放入当前槽位了
	if wheel.slotIdx++; wheel.slotIdx >= wheel.size() {
		wheel.slotIdx = 0
		isCascade = true
	}
	node := slot.next
	slot.next, slot.prev = slot, slot //清空当前格子
	for node != slot {
		tmp := node
		node = node.next
		if tmp.timeDead <= timenow {
			self._AddToReadyNode(tmp)
		} else {
			//【Bug】加新Node，须加到其它槽位，本槽位已扫过(失效，等一整轮才会再扫到)
			self._AddTimerNode(int(tmp.timeDead-timenow), tmp)
		}
	}
	if isCascade {
		self._Cascade(wheelIdx+1, timenow)
	}
}
func (self *timeWheel) _AddToReadyNode(node *TimeNode) {
	node.prev = self.readyNode.prev
	node.prev.next = node
	node.next = &self.readyNode
	self.readyNode.prev = node
}
func (self *timeWheel) _DoTimeOutCallBack() {
	node := self.readyNode.next
	for node != &self.readyNode {
		tmp := node
		node = node.next
		tmp._Callback()
	}
	self.readyNode.next = &self.readyNode
	self.readyNode.prev = &self.readyNode
}

// ------------------------------------------------------------
// 多线程写，单线程读
type SafeMgr struct {
	timeWheel
	queue safe.SafeQueue
}
type obj struct {
	ptr   *TimeNode
	delay int
}

func (self *SafeMgr) Init(cap uint32) {
	self.queue.Init(cap)
	self.timeWheel.Init()
}
func (self *SafeMgr) Refresh(timelapse int, timenow int64) {
	for {
		if v, ok, _ := self.queue.Get(); ok {
			if v := v.(obj); v.delay >= 0 {
				self.timeWheel._AddTimerNode(v.delay, v.ptr)
			} else {
				self.timeWheel.DelTimer(v.ptr)
			}
		} else {
			break
		}
	}
	self.timeWheel.Refresh(timelapse, timenow)
}
func (self *SafeMgr) DelTimer(p *TimeNode) bool {
	ok, _ := self.queue.Put(obj{p, -1})
	if !ok {
		gamelog.Error("timer del fail")
	}
	return ok
}
func (self *SafeMgr) AddTimerSec(cb func(), delay, cd, total float32) *TimeNode {
	p := newNode(cb, delay, cd, total)
	if ok, _ := self.queue.Put(obj{p, int(delay * 1000)}); ok {
		return p
	}
	gamelog.Error("timer add fail")
	return nil
}
