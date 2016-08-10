package gamelog

import (
	"os"
	"time"
)

type TBinaryLog struct {
	file *os.File
}

func NewBinaryLog(name string) *TBinaryLog {
	var err error = nil
	timeStr := time.Now().Format("20060102_150405")
	logFileName := g_logDir + name + "_" + timeStr + ".blog"

	log := new(TBinaryLog)
	log.file, err = os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		Error("BinaryLog OpenFile:%s", err.Error())
		return nil
	}
	return log
}
func (self *TBinaryLog) Close() {
	self.file.Close()
}
func (self *TBinaryLog) Write(data1, data2 [][]byte) {
	for _, v := range data1 {
		self.file.Write(v)
	}
	for _, v := range data2 {
		self.file.Write(v)
	}
}
