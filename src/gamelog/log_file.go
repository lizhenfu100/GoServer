package gamelog

import (
	"common/file"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	Lv_Debug = iota
	Lv_Track //用于外网问题排查，GM调整日志级别
	Lv_Info
	Lv_Warn
	Lv_Error
	Lv_Fatal
	Change_File_CD = time.Hour * 24
)

var (
	g_logDir   = file.GetExeDir() + "/log/"
	g_logger   *log.Logger
	g_logfile  *os.File //用以关闭旧文件
	g_level    = Lv_Debug
	g_levelStr = []string{
		"[D] ",
		"[T] ",
		"[I] ",
		"[W] ",
		"[E] ",
		"[F] ",
	}
)

func InitFileLog(file *os.File) {
	if g_logger == nil {
		g_logfile = file
		g_logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		g_logfile.Close()
		g_logfile = file
		g_logger.SetOutput(file)
	}
	if g_logger == nil {
		panic("InitFileLog log.New == nil")
		return
	}
}
func SetLevel(l int) {
	if l > Lv_Fatal || l < Lv_Debug {
		g_level = Lv_Debug
	} else {
		g_level = l
	}
}
func _log(lv int, format string, v ...interface{}) {
	if lv < g_level {
		return
	}
	str := fmt.Sprintf(g_levelStr[lv]+format, v...)
	g_logger.Output(3, str)
}

func Debug(format string, v ...interface{}) { _log(Lv_Debug, format, v...) }
func Track(format string, v ...interface{}) { _log(Lv_Track, format, v...) }
func Info(format string, v ...interface{})  { _log(Lv_Info, format, v...) }
func Warn(format string, v ...interface{})  { _log(Lv_Warn, format, v...) }
func Error(format string, v ...interface{}) { _log(Lv_Error, format, v...) }
func Fatal(format string, v ...interface{}) {
	_log(Lv_Fatal, format, v...)
	panic(fmt.Sprintf(format, v...))
}

// -------------------------------------
// 每隔一段时间，更换输出文件
func AutoChangeFile(name string) {
	for range time.Tick(Change_File_CD) {
		file.DelExpired(g_logDir, name, 30) //删除30天前的记录
		timeStr := time.Now().Format("20060102_150405")
		if f, err := file.CreateFile(g_logDir, name+timeStr+".log", os.O_WRONLY); err == nil {
			InitFileLog(f)
		}
	}
}
