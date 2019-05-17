package console

import (
	"common"
	"common/tool/wechat"
	"gamelog"
	"nets/tcp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

type cmdFunc func(args []string)

func RegCmd(key string, f cmdFunc)                          { g_cmds[key] = f }
func _Rpc_gm_cmd1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_gm_cmd2(req, ack) }
func _Rpc_gm_cmd2(req, ack *common.NetPack) {
	cmd := req.ReadString()

	args := strings.Split(cmd, " ")
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover HandleCmd\n%v: %s", r, debug.Stack())
		}
	}()
	if cmd, ok := g_cmds[args[0]]; ok {
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
	"wechat":  cmd_wechat,
}

// ------------------------------------------------------------
//! 命令行函数
func cmd_logLv(args []string) {
	lv, _ := strconv.Atoi(args[1])
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
	n, _ := strconv.Atoi(args[1])
	runtime.GOMAXPROCS(n)
	gamelog.Info("cpu cnt(%d) use(%d)", runtime.NumCPU(), runtime.GOMAXPROCS(0))
}
func cmd_wechat(args []string) {
	wechat.SendMsg(args[1])
}
