package safe

import (
	"fmt"
	"runtime/debug"
	"sync"
)

// 放开容量限制的channel，添加不阻塞，接收可能阻塞；传入nil表示退出
type Pipe struct {
	sync.Mutex
	cond sync.Cond
	list []interface{}
}

func (self *Pipe) Init() {
	self.cond.L = &self.Mutex
}
func (self *Pipe) Add(msg interface{}) {
	self.Lock()
	self.list = append(self.list, msg)
	self.Unlock()
	self.cond.Signal()
}
func (self *Pipe) MoveToStack(retList *[]interface{}) {
	self.Lock()
	for len(self.list) == 0 {
		self.cond.Wait() //无数据，阻塞
	}
	for _, v := range self.list {
		*retList = append(*retList, v)
		if v == nil {
			break //stop loop flag
		}
	}
	self.list = self.list[:0]
	self.Unlock()
}

// ------------------------------------------------------------
type EventQueue struct {
	Pipe
	cond sync.WaitGroup
}

func (self *EventQueue) Init() {
	self.Pipe.Init()
}
func (self *EventQueue) Add(cb func()) {
	if cb != nil {
		self.Pipe.Add(cb)
	}
}
func (self *EventQueue) Loop() {
	self.cond.Add(1)
	var msgs []interface{}
LOOP1:
	for {
		msgs = msgs[:0]
		self.MoveToStack(&msgs)
		for _, msg := range msgs {
			switch v := msg.(type) {
			case func():
				safeCall(v)
			case nil:
				break LOOP1
			default:
				fmt.Printf("unexpected type %T\n", v)
			}
		}
	}
	self.cond.Done()
}
func (self *EventQueue) WaitStop() { self.Pipe.Add(nil); self.cond.Wait() }

func safeCall(cb func()) {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
		}
	}()
	cb()
}
