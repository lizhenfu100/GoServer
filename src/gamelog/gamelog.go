package gamelog

import (
	"common"
	"conf"
	"os"
	"time"
)

// -------------------------------------
//
func InitLogger(name string) {
	if conf.IsDebug {
		InitFileLog(os.Stdout)
	} else {
		timeStr := time.Now().Format("20060102_150405")
		file, err := common.CreateFile(g_logDir, name+timeStr+".log")
		if err != nil {
			panic("CreateFile error : " + err.Error())
		}
		InitFileLog(file)
		// _initAsyncLog(name)
	}
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
