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

func newFile(name string) *os.File {
	timeStr := time.Now().Format("20060102_150405")
	if f, err := file.CreateFile(g_logDir, name+timeStr+".log", os.O_WRONLY); err == nil {
		if g_logfile != nil {
			g_logfile.Close()
		}
		g_logfile = f
		return f
	}
	return nil
}

// 每隔一段时间，更换输出文件
func AutoChangeFile(name string, p *AsyncLog) {
	for range time.Tick(24 * time.Hour) {
		file.DelExpired(g_logDir, name, 30)
		if f := newFile(name); f != nil {
			p.wr.Store(f)
		}
	}
}
