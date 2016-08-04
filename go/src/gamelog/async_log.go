/***********************************************************************
* @ 异步日志
* @ brief
	1、前端Append()接口，用以输入数据，buf被写满时触发后台writeLoop

	2、后台writeLoop平时阻塞在"<-self.awakeChan"处，等待chan写操作的唤醒

	3、timeOutWrite为了及时记log

	4、若强杀Log进程，可能buf中的数据还没被写

* @ race condition
	1、"go chan"内部也是锁实现的，chan操作不要放在临界区
		否则就锁中套锁了，极易出问题

	2、比如连续两次触发buf被写满，第二次的chan会阻塞，挂起Append()的线程
		若chan位于临界区内则还占用着Mutex
		后台writeLoop被唤醒时，同样要访问临界区，就被挂起了
		然后两线程此时就都挂着咯~

* @ author zhoumf
* @ date 2016-8-4
***********************************************************************/
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

func NewAsyncLog(bufSize int) *AsyncLog {
	log := new(AsyncLog)
	log.curBuf = make([][]byte, 0, bufSize)
	log.spareBuf = make([][]byte, 0, bufSize)
	log.awakeChan = make(chan bool)
	go log._writeLoop(bufSize)
	go log._timeOutWrite()
	return log
}

//如果写得非常快，瞬间把两片buf都写满了，会阻塞在awakeChan处，等writeLoop写完log即恢复
//两片buf的好处：在当前线程即可交换，不用等到后台writeLoop唤醒
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

func (self *AsyncLog) _writeLoop(bufSize int) {
	bufToWrite1 := make([][]byte, 0, bufSize)
	bufToWrite2 := make([][]byte, 0, bufSize)
	for {
		<-self.awakeChan //没人写数据即阻塞：超时/buf写满，唤起【这句不能放在临界区，否则死锁】

		self.Lock()
		{
			//此时bufToWrite为空，交换
			_swapBuf(&bufToWrite1, &self.spareBuf)
			_swapBuf(&bufToWrite2, &self.curBuf)
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
