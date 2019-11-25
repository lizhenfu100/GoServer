package gamelog

import (
	"common/file"
	"os"
	"time"
)

var (
	g_logDir  = file.GetExeDir() + "/log/"
	g_logfile *os.File //用以关闭旧文件
)

func InitFileLog(name string) {
	if g_logfile != nil {
		g_logfile.Close()
	}
	g_logfile = getFile(name)
	g_log.SetOutput(g_logfile)

	if g_logfile == nil {
		panic("InitFileLog nil")
		return
	}
	go func() {
		file.DelExpired(g_logDir, name, 30) //删除30天前的记录
		if f := getFile(name); f != nil {   //更换输出文件
			g_logfile.Close()
			g_logfile = f
			g_log.SetOutput(f)
		}
	}()
}
func getFile(name string) *os.File {
	timeStr := time.Now().Format("20060102_150405")
	if f, err := file.CreateFile(g_logDir, name+timeStr+".log", os.O_WRONLY); err == nil {
		return f
	}
	return nil
}

// 每隔一段时间，更换输出文件
func AutoChangeFile(name string, p *AsyncLog) {
	for range time.Tick(24 * time.Hour) {
		file.DelExpired(g_logDir, name, 30) //删除30天前的记录
		if f := getFile(name); f != nil {
			g_logfile.Close()
			g_logfile = f
			p.wr.Set(f)
		}
	}
}
