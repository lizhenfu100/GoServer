package gamelog

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var G_CurPath string

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

var (
	g_logDir = GetCurrPath() + "log\\"
)

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
	InitBinaryLog()
}
