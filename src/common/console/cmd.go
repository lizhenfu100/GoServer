package console

import (
	"common"
	"common/assert"
	"fmt"
	"gamelog"
	"nets/tcp"
	"runtime"
	"strconv"
	"strings"
)

type cmdFunc func(args []string)

func RegCmd(key string, f cmdFunc) {
	if _, ok := g_cmds[key]; ok {
		assert.True(false)
		gamelog.Error("RegCmd repeat: " + key)
		return
	}
	g_cmds[key] = f
}
func _Rpc_gm_cmd1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_gm_cmd2(req, ack) }
func _Rpc_gm_cmd2(req, ack *common.NetPack) {
	cmd := req.ReadString()
	args_ := req.ReadString()
	args := strings.Split(args_, " ")

	defer func() {
		if r := recover(); r != nil {
			ack.WriteString(fmt.Sprintf("%v", r))
		}
	}()
	if cmd, ok := g_cmds[cmd]; ok {
		cmd(args)
		ack.WriteString("ok")
	} else {
		ack.WriteString("none cmd")
	}
}

var g_cmds = map[string]cmdFunc{ //Notice：注意下列函数的线程安全性
	"loglv":   cmd_logLv,
	"gc":      cmd_gc,
	"routine": cmd_routine,
	"cpu":     cmd_cpu,
	"setcpu":  cmd_setcpu,
}

// ------------------------------------------------------------
//! 命令行函数
func cmd_logLv(args []string) {
	lv, _ := strconv.Atoi(args[0])
	gamelog.SetLevel(lv)
	gamelog.Info("SetLogLv: %d", lv)
}
func cmd_gc(args []string) {
	runtime.GC()
	gamelog.Info("GC finished")
}
func cmd_routine(args []string) {
	gamelog.Info("Current number of goroutines: %d", runtime.NumGoroutine())
}
func cmd_cpu(args []string) {
	gamelog.Info("cpu cnt(%d) use(%d)", runtime.NumCPU(), runtime.GOMAXPROCS(0))
}
func cmd_setcpu(args []string) {
	n, _ := strconv.Atoi(args[0])
	runtime.GOMAXPROCS(n)
	gamelog.Info("cpu cnt(%d) use(%d)", runtime.NumCPU(), runtime.GOMAXPROCS(0))
}
