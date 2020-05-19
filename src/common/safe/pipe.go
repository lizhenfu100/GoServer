package safe

import "sync"

type Pipe struct { //非阻塞，单消费者
	sync.Mutex
	multi [2][]interface{}
	cur   []interface{}
	cycle uint8
}
type Chan struct { //阻塞，单消费者
	sync.Cond
	Pipe
	IsStop bool
}

func (p *Pipe) Init(cap uint32) {
	for i := 0; i < len(p.multi); i++ {
		p.multi[i] = make([]interface{}, 0, cap)
	}
	p.cur = p.multi[0]
}
func (p *Pipe) Add(v interface{}) {
	p.Lock()
	p.cur = append(p.cur, v)
	p.Unlock()
}
func (p *Pipe) Get() (ret []interface{}) {
	p.Lock()
	ret = p._get()
	p.Unlock()
	return
}
func (p *Pipe) _get() (ret []interface{}) {
	ret = p.cur //给消费者
	p.cycle = (p.cycle + 1) % uint8(len(p.multi))
	p.cur = p.multi[p.cycle] //生产者指向新队列
	p.cur = p.cur[:0]
	return ret
}
func (p *Chan) Init(cap uint32) {
	p.Pipe.Init(cap)
	p.Cond.L = &p.Mutex
}
func (p *Chan) Add(v interface{}) {
	p.Lock()
	p.cur = append(p.cur, v)
	p.Unlock()
	p.Signal()
}
func (p *Chan) Get() (ret []interface{}) {
	p.Lock()
	for len(p.cur) == 0 && !p.IsStop {
		p.Wait()
	}
	ret = p._get()
	p.Unlock()
	return
}
func (p *Chan) Stop() {
	p.Lock()
	p.IsStop = true
	p.Unlock()
	p.Signal()
}

// ------------------------------------------------------------
type ChanByte struct { //阻塞，单消费者
	sync.Mutex
	sync.Cond
	multi  [2][]byte
	Cur    []byte
	cycle  uint8
	IsStop bool
}

func (p *ChanByte) Init(cap uint32) {
	for i := 0; i < len(p.multi); i++ {
		p.multi[i] = make([]byte, 0, cap)
	}
	p.Cur = p.multi[0]
	p.Cond.L = &p.Mutex
}
func (p *ChanByte) Add(v []byte) {
	p.Lock()
	p.Cur = append(p.Cur, v...)
	p.Unlock()
	p.Signal()
}
func (p *ChanByte) Get() (ret []byte) {
	p.Lock()
	for len(p.Cur) == 0 && !p.IsStop {
		p.Wait()
	}
	ret = p.Cur
	p.cycle = (p.cycle + 1) % uint8(len(p.multi))
	p.Cur = p.multi[p.cycle]
	p.Cur = p.Cur[:0]
	p.Unlock()
	return
}
func (p *ChanByte) Stop() {
	p.Lock()
	p.IsStop = true
	p.Unlock()
	p.Signal()
}
