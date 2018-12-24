package console

import (
	"gamelog"
	"os"
	"os/signal"
	"syscall"
)

func RegShutdown(f cmdFunc) { RegCmd("shutdown", f) }

func SIGTERM() { // 监控进程终止信号
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM)
	s := <-c //阻塞直至有信号传入

	gamelog.Info("get signal: %v", s)
	if f, ok := g_cmds["shutdown"]; ok { //模块注册的关服函数
		f(nil)
	} else {
		os.Exit(0)
	}
}
