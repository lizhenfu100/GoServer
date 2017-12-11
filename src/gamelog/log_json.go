package gamelog

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

type M map[string]interface{}
type TJsonLog struct {
	file *os.File
	wr   *bufio.Writer
	json *json.Encoder
}

func NewJsonLog(name string) *TJsonLog {
	var err error = nil
	timeStr := time.Now().Format("20060102_150405")
	fullName := g_logDir + name + "_" + timeStr + ".jlog"

	log := new(TJsonLog)
	log.file, err = os.OpenFile(fullName, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Error("JsonLog OpenFile:%v", err.Error())
		return nil
	}
	log.wr = bufio.NewWriterSize(log.file, 1024)
	log.json = json.NewEncoder(log.wr)

	return log
}
func (self *TJsonLog) Close() {
	self.wr.Flush()
	self.file.Close()
}
func (self *TJsonLog) Flush() {
	self.wr.Flush()
}

// Append(gamelog.M{"a":1, "b":Struct{233,"zhoumf"}})
func (self *TJsonLog) Append(data M) {
	self.json.Encode(data)
}
