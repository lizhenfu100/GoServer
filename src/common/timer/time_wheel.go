/***********************************************************************
* @ 【单线程】时间轮
* @ brief

* @ author zhoumf
* @ date 2017-5-11
***********************************************************************/
package timer

import (
	"common/assert"
	"time"
)

const (
	kWheelNum    = 5
	kTimeTickLen = 25 //一格的刻度 ms
)

var (
	kWheelBit  = [...]uint{8, 6, 6, 6, 5} //每一级的槽位数量；用了累计位移，总和不可超32
	kWheelSize [kWheelNum]uint
	kWheelCap  [kWheelNum]uint
	G_TimerMgr = _NewTimerMgr()
)

// ------------------------------------------------------------
// node
type timeNode struct {
	prev     *timeNode
	next     *timeNode
	timeDead int64
	interval int
	total    int
	callback func()
}

func newTimeNode() *timeNode {
	ret := new(timeNode)
	ret.init()
	return ret
}
func (self *timeNode) init() {
	self.prev = self
	self.next = self //circle
}
func (self *timeNode) _Callback() {
	isJoin := false
	if self.total < 0 {
		isJoin = true
	} else if self.total -= self.interval; self.total > 0 {
		isJoin = true
	}
	if isJoin {
		self.timeDead += int64(self.interval)
		G_TimerMgr._AddTimerNode(self.interval, self)
	}
	self.callback()
}

// ------------------------------------------------------------
// wheel
type stWheel struct {
	//每个slot维护的node链表为一个环，如此可以简化插入删除的操作
	//slot.next为node链表中第一个节点，prev为node的最后一个节点
	slots   []timeNode
	slotIdx uint
}

func (self *stWheel) init(size uint) {
	self.slots = make([]timeNode, size)
	for i := uint(0); i < size; i++ {
		self.slots[i].init()
	}
}
func (self *stWheel) GetCurSlot() *timeNode {
	return &self.slots[self.slotIdx]
}
func (self *stWheel) size() uint { return uint(len(self.slots)) }

// ------------------------------------------------------------
// manager
type timerMgr struct {
	wheels     [kWheelNum]stWheel
	readyNode  timeNode
	timeElapse int
}

func _NewTimerMgr() *timerMgr {
	assert.True(len(kWheelBit) == kWheelNum)
	ret := new(timerMgr)
	ret.readyNode.init()
	for i := 0; i < kWheelNum; i++ {
		var wheelCap uint
		for j := 0; j <= i; j++ {
			wheelCap += kWheelBit[j]
		}
		assert.True(wheelCap < 32)
		kWheelCap[i] = 1 << wheelCap
		kWheelSize[i] = 1 << kWheelBit[i]
		ret.wheels[i].init(kWheelSize[i])
	}
	return ret
}
func (self *timerMgr) Refresh(timeElapse int, timenow int64) {
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
func (self *timerMgr) AddTimerSec(callback func(), delay, cd, total float32) {
	node := newTimeNode()
	node.callback = callback
	node.interval = int(cd * 1000)
	node.total = int(total * 1000)
	node.timeDead = time.Now().UnixNano()/int64(time.Millisecond) + int64(delay*1000)
	self._AddTimerNode(int(delay*1000), node)
}
func (self *timerMgr) _AddTimerNode(msec int, node *timeNode) {
	var slot *timeNode
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
func (self *timerMgr) DelTimer(node *timeNode) {
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	node.prev, node.next = nil, nil
}
func (self *timerMgr) _Cascade(wheelIdx int, timenow int64) {
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
func (self *timerMgr) _AddToReadyNode(node *timeNode) {
	node.prev = self.readyNode.prev
	node.prev.next = node
	node.next = &self.readyNode
	self.readyNode.prev = node
}
func (self *timerMgr) _DoTimeOutCallBack() {
	node := self.readyNode.next
	for node != &self.readyNode {
		tmp := node
		node = node.next
		tmp._Callback()
	}
	self.readyNode.next = &self.readyNode
	self.readyNode.prev = &self.readyNode
}
