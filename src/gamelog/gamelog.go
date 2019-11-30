package gamelog

import (
	"common/assert"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func InitLogger(name string) {
	if assert.IsDebug {
		g_log.SetOutput(os.Stdout)
		return
	}
	SetLevel(Lv_Info)
	p := NewAsyncLog(10240, newFile(name))
	g_log.SetOutput(p)
	go AutoChangeFile(name, p)
}

// ------------------------------------------------------------
const (
	Lv_Debug = iota
	Lv_Track //用于外网问题排查，GM调整日志级别
	Lv_Info
	Lv_Warn
	Lv_Error
	Lv_Fatal
	Flush_Interval = 180 * time.Second //隔几秒刷次log
)

var (
	g_log      = log.New(nil, "", log.Ldate|log.Ltime|log.Lshortfile)
	g_level    = Lv_Debug
	g_levelStr = []string{
		"[D] ",
		"[T] ",
		"[I] ",
		"[W] ",
		"[E] ",
		"[E] ",
	}
)

func SetLevel(l int) {
	if l > Lv_Fatal || l < Lv_Debug {
		g_level = Lv_Debug
	} else {
		g_level = l
	}
}
func _log(lv int, format string, v ...interface{}) {
	if lv < g_level {
		return
	}
	str := fmt.Sprintf(g_levelStr[lv]+format, v...)
	g_log.Output(3, str)

}
func Debug(format string, v ...interface{}) { _log(Lv_Debug, format, v...) }
func Track(format string, v ...interface{}) { _log(Lv_Track, format, v...) }
func Info(format string, v ...interface{})  { _log(Lv_Info, format, v...) }
func Warn(format string, v ...interface{})  { _log(Lv_Warn, format, v...) }
func Error(format string, v ...interface{}) { _log(Lv_Error, format, v...) }
func Fatal(format string, v ...interface{}) {
	_log(Lv_Fatal, format, v...)
	panic(fmt.Sprintf(format, v...))
}

// ------------------------------------------------------------
type AsyncLog struct {
	sync.Mutex
	cond    sync.Cond
	curBuf  []byte
	nextBuf []byte
	bufs    [][]byte
	wr      atomic.Value //io.Writer
	isStop  bool
}

func NewAsyncLog(cap int, wr io.Writer) *AsyncLog {
	self := &AsyncLog{
		curBuf:  make([]byte, 0, cap),
		nextBuf: make([]byte, 0, cap),
	}
	self.wr.Store(wr)
	self.cond.L = &self.Mutex
	go self._writeLoop(cap)
	go func() {
		for range time.Tick(Flush_Interval) { //周期性唤醒写线程
			self.cond.Signal()
		}
	}()
	return self
}
func (self *AsyncLog) Write(p []byte) (int, error) {
	self.Lock()
	nCap := cap(self.curBuf)
	self.curBuf = append(self.curBuf, p...)
	if len(self.curBuf) >= nCap {
		self.bufs = append(self.bufs, self.curBuf)
		if self.nextBuf != nil {
			_move(&self.curBuf, &self.nextBuf)
		} else {
			self.curBuf = make([]byte, 0, nCap)
		}
		self.cond.Signal()
	}
	self.Unlock()
	return len(p), nil
}
func (self *AsyncLog) _writeLoop(cap int) {
	spareBuf1 := make([]byte, 0, cap)
	spareBuf2 := make([]byte, 0, cap)
	bufs := make([][]byte, 0, 8)
	for !self.isStop || !self.isNil() {
		self.Lock()
		if len(self.bufs) == 0 {
			self.cond.Wait()
		}
		self.bufs = append(self.bufs, self.curBuf)
		_move(&self.curBuf, &spareBuf1)
		_swap(&self.bufs, &bufs)
		if self.nextBuf == nil {
			_move(&self.nextBuf, &spareBuf2)
		}
		self.Unlock()

		wr := self.wr.Load().(io.Writer) //copy to stack
		for _, v := range bufs {
			wr.Write(v)
		}
		if spareBuf1 == nil {
			_move(&spareBuf1, &bufs[0])
			spareBuf1 = spareBuf1[:0]
		}
		if spareBuf2 == nil {
			_move(&spareBuf2, &bufs[1])
			spareBuf2 = spareBuf2[:0]
		}
		bufs = bufs[:0]
	}
}
func _move(rhs, lhs *[]byte)   { *rhs = *lhs; *lhs = nil }
func _swap(rhs, lhs *[][]byte) { temp := *rhs; *rhs = *lhs; *lhs = temp }
func (self *AsyncLog) Stop() {
	self.isStop = true
	self.cond.Signal()
}
func (self *AsyncLog) isNil() bool {
	self.Lock()
	ret := len(self.curBuf) == 0 && len(self.bufs) == 0
	self.Unlock()
	return ret
}
