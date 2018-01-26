/***********************************************************************
* @ 异步日志
* @ brief
	1、前端Append()接口，用以输入数据，buf被写满时触发后台writeLoop

	2、后台writeLoop平时阻塞在"self.cond.Wait()"处，等待唤醒

	3、timeOutWrite为了及时记log

	4、若强杀Log进程，可能buf中的数据还没被写

	5、NewAsyncLog()后立即调Appen()，可能"go _writeLoop"还没来得及启动

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
	Flush_Interval = time.Second * 15
)

type Writer interface {
	Write(data1, data2 [][]byte)
}
type AsyncLog struct {
	sync.Mutex
	cond     *sync.Cond
	curBuf   [][]byte
	spareBuf [][]byte
	wr       Writer
}

func NewAsyncLog(bufSize int, wr Writer) *AsyncLog {
	log := new(AsyncLog)
	log.cond = sync.NewCond(&log.Mutex)
	log.curBuf = make([][]byte, 0, bufSize)
	log.spareBuf = make([][]byte, 0, bufSize)
	log.wr = wr
	go log._writeLoop(bufSize)
	go log._timeOutWrite()
	return log
}

//如果写得非常快，瞬间把两片buf都写满了，会阻塞在awakeChan处，等writeLoop写完log即恢复
//两片buf的好处：在当前线程即可交换，不用等到后台writeLoop唤醒
func (self *AsyncLog) Append(pdata []byte) {
	self.Lock()
	{
		self.curBuf = append(self.curBuf, pdata)
		if len(self.curBuf) == cap(self.curBuf) {
			_swapBuf(&self.curBuf, &self.spareBuf)
			self.cond.Signal()
		}
	}
	self.Unlock()
}

func (self *AsyncLog) _writeLoop(bufSize int) {
	bufToWrite1 := make([][]byte, 0, bufSize)
	bufToWrite2 := make([][]byte, 0, bufSize)
	for {
		self.Lock()
		{
			self.cond.Wait() //Notice：必须在临近区内

			//此时bufToWrite为空，交换
			_swapBuf(&bufToWrite1, &self.spareBuf)
			_swapBuf(&bufToWrite2, &self.curBuf)
		}
		self.Unlock()

		//将bufToWrite中的数据全写进log，并清空
		self.wr.Write(bufToWrite1, bufToWrite2)
		bufToWrite1 = bufToWrite1[:0]
		bufToWrite2 = bufToWrite2[:0]
	}
}
func (self *AsyncLog) _timeOutWrite() {
	for range time.Tick(Flush_Interval) {
		self.cond.Signal()
	}
}
func _swapBuf(rhs, lhs *[][]byte) {
	temp := *rhs
	*rhs = *lhs
	*lhs = temp
}
