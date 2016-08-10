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

	g_logDir = GetCurrPath() + "log\\"

	G_AsyncLog  *AsyncLog
	g_binaryLog *TBinaryLog
	g_mysqlLog  *TMysqlLog
)

func GetCurrPath() string {
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

func InitLogger(name string, bScreen bool) {
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

	InitDebugLog(logFileName, bScreen)

	_initAsyncLog(name)

}
func _initAsyncLog(name string) {
	G_AsyncLog = NewAsyncLog(1024, _doWriteBinaryLog)

	g_binaryLog = NewBinaryLog("logsvr")
	g_mysqlLog = NewMysqlLog()
	if g_binaryLog == nil || g_mysqlLog == nil {
		panic("New Log fail!")
		return
	}
}
func _doWriteBinaryLog(data1, data2 [][]byte) {
	g_binaryLog.Write(data1, data2)
}
func _doWriteMysqlLog(data1, data2 [][]byte) {
	g_mysqlLog.Write(data1, data2)
}
