package gamelog

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	G_CurPath string

	g_logDir = GetExePath() + "log\\"

	G_AsyncLog *AsyncLog
)

//////////////////////////////////////////////////////////////////////
// 辅助函数
func GetExePath() string {
	if len(G_CurPath) <= 0 {
		file, _ := exec.LookPath(os.Args[0])
		G_CurPath, _ = filepath.Abs(file)
		G_CurPath = string(G_CurPath[0 : 1+strings.LastIndex(G_CurPath, "\\")])
	}
	return G_CurPath
}
func IsDirExists(path string) bool {
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
	if !IsDirExists(g_logDir) {
		err = os.MkdirAll(g_logDir, os.ModePerm)
	}
	if err != nil {
		panic("InitLogger error : " + err.Error())
		return
	}

	timeStr := time.Now().Format("20060102_150405")
	logFileName := g_logDir + name + "_" + timeStr + ".log"

	InitDebugLog(logFileName)

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
