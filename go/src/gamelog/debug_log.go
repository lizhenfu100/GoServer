package gamelog

import (
	"fmt"
	"log"
	"os"
)

const (
	InfoLevel = iota
	WarnLevel
	ErrorLevel
	FatalLevel
)

var (
	g_logger         *log.Logger
	g_isOutputScreen = true
	g_level          = InfoLevel
)

func InitDebugLog(logFileName string) {
	file, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE, 0)
	if err != nil {
		panic("InitLogger error : " + err.Error())
		return
	}

	g_logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	if g_logger == nil {
		panic("InitLogger error : " + err.Error())
		return
	}
}
func GetLevel() int {
	return g_level
}
func SetLevel(l int) {
	if l > FatalLevel || l < InfoLevel {
		g_level = InfoLevel
	} else {
		g_level = l
	}
}

func Warn(format string, v ...interface{}) {
	if g_level <= WarnLevel {
		str := "[W] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(2, str)

		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
func Error(format string, v ...interface{}) {
	if g_level <= ErrorLevel {
		str := "[E] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(2, str)
		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
func Error3(format string, v ...interface{}) {
	if g_level <= ErrorLevel {
		str := "[E] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(3, str)

		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
func Info(format string, v ...interface{}) {
	if g_level <= InfoLevel {
		str := "[I] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(2, str)

		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
func Fatal(format string, v ...interface{}) {
	if g_level <= FatalLevel {
		str := "[F] " + format
		str = fmt.Sprintf(str, v...)
		g_logger.Output(4, str)

		if g_isOutputScreen {
			fmt.Println(str)
		}
	}
}
