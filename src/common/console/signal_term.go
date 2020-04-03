package console

import (
	"os"
	"os/signal"
	"syscall"
)

var g_shutdown func()

func RegShutdown(f func()) { g_shutdown = f }

func sigTerm() { // 监控进程终止信号
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM)
	_ = <-c //阻塞直至有信号传入

	if g_shutdown != nil {
		g_shutdown()
	} else {
		os.Exit(0)
	}
}
