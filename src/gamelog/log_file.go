package gamelog

import (
	"fmt"
	"io"
	"log"
)

const (
	Lv_Debug = iota
	Lv_Info
	Lv_Warn
	Lv_Error
	Lv_Fatal
)

var (
	g_logger   *log.Logger
	g_level    = Lv_Debug
	g_levelStr = []string{
		"[D] ",
		"[I] ",
		"[W] ",
		"[E] ",
		"[F] ",
	}
)

func InitFileLog(out io.Writer) {
	g_logger = log.New(out, "", log.Ldate|log.Ltime|log.Lshortfile)
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
func Info(format string, v ...interface{})  { _log(Lv_Info, format, v...) }
func Warn(format string, v ...interface{})  { _log(Lv_Warn, format, v...) }
func Error(format string, v ...interface{}) { _log(Lv_Error, format, v...) }
func Fatal(format string, v ...interface{}) {
	_log(Lv_Fatal, format, v...)
	panic(fmt.Sprintf(format, v...))
}
