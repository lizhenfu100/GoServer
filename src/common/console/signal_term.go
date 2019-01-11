package console

import (
	"os"
	"os/signal"
	"syscall"
)

func RegShutdown(f cmdFunc) { RegCmd("shutdown", f) }

func sigTerm() { // 监控进程终止信号
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM)
	_ = <-c //阻塞直至有信号传入

	if f, ok := g_cmds["shutdown"]; ok { //模块注册的关服函数
		f(nil)
	} else {
		os.Exit(0)
	}
}
