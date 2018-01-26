package gamelog

import (
	"common/file"
	"conf"
	"os"
	"time"
)

// -------------------------------------
//
func InitLogger(name string) {
	if conf.IsDebug {
		InitFileLog(os.Stdout)
		return
	}
	timeStr := time.Now().Format("20060102_150405")
	if f, err := file.CreateFile(g_logDir, name+timeStr+".log", os.O_WRONLY); err == nil {
		InitFileLog(f)
		// _initAsyncLog(f)
	} else {
		panic("CreateFile error : " + err.Error())
	}
	SetLevel(Lv_Info)
	go AutoChangeFile(name)
}

// -------------------------------------
// 异步日志
// var G_AsyncLog *AsyncLog

// func _initAsyncLog(name string) {
// 	G_AsyncLog = NewAsyncLog(1024, NewBinaryLog("logsvr"))
// 	// G_AsyncLog = NewAsyncLog(1024, NewMysqlLog("logsvr"))

// 	if G_AsyncLog == nil {
// 		panic("New Log fail!")
// 		return
// 	}
// }
// func AppendAsyncLog(data []byte) {
// 	G_AsyncLog.Append(data)
// }
