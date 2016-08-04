package gamelog

import (
	"sync"
	"time"
)

const (
	Flush_Interval = 15 //间隔几秒写一次log
)

type AsyncLog struct {
	sync.Mutex
	curBuf    [][]byte
	spareBuf  [][]byte
	awakeChan chan bool //chan要make初始化才能用~o(╯□╰)o
}

func NewAsyncLog() *AsyncLog {
	log := new(AsyncLog)
	log.curBuf = make([][]byte, 0, 1024)
	log.spareBuf = make([][]byte, 0, 1024)
	log.awakeChan = make(chan bool)
	go log._writeLoop()
	go log._timeOutWrite()
	return log
}

//如果写得非常快，瞬间把两片buf都写满了，会阻塞在awakeChan处，等writeLoop写完log即恢复
//两片buf的好处：在当前线程即可交互，不用等到后台writeLoop唤醒
func (self *AsyncLog) Append(pdata []byte) {
	isAwakenWriteLoop := false
	self.Lock()
	{
		self.curBuf = append(self.curBuf, pdata)
		if len(self.curBuf) == cap(self.curBuf) {
			_swapBuf(&self.curBuf, &self.spareBuf)
			isAwakenWriteLoop = true
		}
	}
	self.Unlock()

	if isAwakenWriteLoop {
		self.awakeChan <- true //Notice：不能放在临界区
	}
}

func (self *AsyncLog) _writeLoop() {
	bufToWrite1 := make([][]byte, 0, 1024)
	bufToWrite2 := make([][]byte, 0, 1024)
	for {
		<-self.awakeChan //没人写数据即阻塞：超时/buf写满，唤起【这句不能放在临界区，否则死锁】

		self.Lock()
		{
			//此时bufToWrite为空，交换
			_swapBuf(&self.spareBuf, &bufToWrite1)
			_swapBuf(&self.curBuf, &bufToWrite2)
		}
		self.Unlock()

		//将bufToWrite中的数据全写进log，并清空
		WriteBinaryLog(bufToWrite1, bufToWrite2)
		_clearBuf(&bufToWrite1)
		_clearBuf(&bufToWrite2)
	}
}
func (self *AsyncLog) _timeOutWrite() {
	for {
		time.Sleep(Flush_Interval * time.Second)
		self.awakeChan <- true
	}
}
func _swapBuf(rhs, lhs *[][]byte) {
	temp := *rhs
	*rhs = *lhs
	*lhs = temp
}
func _clearBuf(p *[][]byte) {
	*p = append((*p)[:0], [][]byte{}...)
}
