package gamelog

import (
	"os"
	"path/filepath"
	"time"
)

var (
	g_logDir = GetExeDir() + "log\\"

	G_AsyncLog *AsyncLog
)

//////////////////////////////////////////////////////////////////////
// 辅助函数
func GetExeDir() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir + "\\"
}
func IsDirExist(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
	return true
}

//////////////////////////////////////////////////////////////////////
//
func InitLogger(name string) {
	var err error = nil
	if !IsDirExist(g_logDir) {
		err = os.MkdirAll(g_logDir, os.ModePerm)
	}
	if err != nil {
		panic("InitLogger error : " + err.Error())
		return
	}

	timeStr := time.Now().Format("20060102_150405")
	fullName := g_logDir + name + "_" + timeStr + ".log"
	InitDebugLog(fullName)

	_initAsyncLog(name)

}
func _initAsyncLog(name string) {
	G_AsyncLog = NewAsyncLog(1024, NewBinaryLog("logsvr"))
	// G_AsyncLog = NewAsyncLog(1024, NewMysqlLog("logsvr"))

	if G_AsyncLog == nil {
		panic("New Log fail!")
		return
	}
}
func AppendAsyncLog(data []byte) {
	G_AsyncLog.Append(data)
}
