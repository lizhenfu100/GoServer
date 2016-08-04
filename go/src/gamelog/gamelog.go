package gamelog

import (
	"fmt"
	"log"
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

const (
	InfoLevel = iota
	WarnLevel
	ErrorLevel
	FatalLevel

	BinaryLog_Rename_Interval = 60 * 15
)

var g_logger *log.Logger
var g_level = InfoLevel
var g_logFile *os.File
var g_isOutputScreen = true
var g_logDir = GetCurrPath() + "log\\"
var g_BinaryLogTime time.Time

func GetLevel() int {
	return g_level
}
func SetLevel(lv int) {
	if lv > FatalLevel || lv < InfoLevel {
		g_level = InfoLevel
	} else {
		g_level = lv
	}
}

func InitLogger(strModule string, bScreen bool) bool {
	var err error = nil
	if !IsDirExists(g_logDir) {
		err = os.MkdirAll(g_logDir, 0777)
	}

	if err != nil {
		panic("InitLogger error : " + err.Error())
		return false
	}

	timeStr := time.Now().Format("20060102_150405")
	strModule = g_logDir + strModule + "_" + timeStr + ".log"
	g_logFile, err = os.OpenFile(strModule, os.O_WRONLY|os.O_CREATE, 0)
	if err != nil {
		panic("InitLogger error : " + err.Error())
		return false
	}

	g_logger = log.New(g_logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	if g_logger == nil {
		panic("InitLogger error : " + err.Error())
		return false
	}

	g_isOutputScreen = bScreen

	g_BinaryLogTime = time.Now()

	return true
}

func Warn(format string, v ...interface{}) {
	if g_level <= WarnLevel {
		var str string
		str = "[W] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(2, str)

		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
func Error(format string, v ...interface{}) {
	if g_level <= ErrorLevel {
		var str string
		str = "[E] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(2, str)
		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
func Error3(format string, v ...interface{}) {
	if g_level <= ErrorLevel {
		var str string
		str = "[E] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(3, str)
		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
func Info(format string, v ...interface{}) {
	if g_level <= InfoLevel {
		var str string
		str = "[I] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(2, str)

		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
func Fatal(format string, v ...interface{}) {
	if g_level <= FatalLevel {
		var str string
		str = "[F] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(4, str)

		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}

// 二进制文件log
//
func WriteBinaryLog(data1, data2 [][]byte) {
	logFileName := _getBinaryLogName()
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		Error("logsvr OpenFile:%s", err.Error())
		return
	}
	for _, v := range data1 {
		file.Write(v)
	}
	for _, v := range data2 {
		file.Write(v)
	}
}
func _getBinaryLogName() string {
	timeNow := time.Now()
	var timeStr string
	if timeNow.Unix()-g_BinaryLogTime.Unix() >= BinaryLog_Rename_Interval {
		timeStr = timeNow.Format("20060102_150405")
	} else {
		timeStr = g_BinaryLogTime.Format("20060102_150405")
	}
	return g_logDir + "test_logsvr_" + timeStr + ".log"
}
