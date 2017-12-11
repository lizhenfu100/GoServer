package gamelog

import (
	"bufio"
	"os"
	"time"
)

type TBinaryLog struct {
	file *os.File
	wr   *bufio.Writer
}

func NewBinaryLog(name string) *TBinaryLog {
	var err error = nil
	timeStr := time.Now().Format("20060102_150405")
	logFileName := g_logDir + name + "_" + timeStr + ".blog"

	log := new(TBinaryLog)
	log.file, err = os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Error("BinaryLog OpenFile:%s", err.Error())
		return nil
	}
	log.wr = bufio.NewWriterSize(log.file, 1024)

	return log
}
func (self *TBinaryLog) Close() {
	self.wr.Flush()
	self.file.Close()
}

// 用bufio，减少直接file.Write的IO次数
// 实际上是，在bufio内部将data1+data2整合成一个[]byte，再调file.Write(data)
// 用内存拷贝，节省IO次数
func (self *TBinaryLog) Write(data1, data2 [][]byte) {
	for _, v := range data1 {
		self.wr.Write(v)
	}
	for _, v := range data2 {
		self.wr.Write(v)
	}
	self.wr.Flush()
}
