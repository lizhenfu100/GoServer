package console

import (
	"common"
	"conf"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"http"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"tcp"
)

type cmdFunc func(args []string)

var g_cmds = map[string]cmdFunc{ //Notice：注意下列函数的线程安全性
	"loglv":   cmd_LogLv,
	"gc":      cmd_gc,
	"routine": cmd_routine,
	"cpu":     cmd_cpu,
	"setcpu":  cmd_setcpu,
}

func Init() {
	tcp.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_cmd_tcp
	http.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_cmd_http
	tcp.G_HandleFunc[enum.Rpc_log] = _Rpc_log_tcp
	http.G_HandleFunc[enum.Rpc_log] = _Rpc_log_http
	go sigTerm() //监控进程终止信号
}

func _Rpc_log_tcp(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_log_http(req, ack) }
func _Rpc_log_http(req, ack *common.NetPack) {
	log := req.ReadString()
	uuid := req.ReadString()
	version := req.ReadString()
	platform := req.ReadString()
	gamelog.Info("Client Log: %s\nUUID: %s version: %s platform: %s", log, uuid, version, platform)
}

func RegCmd(key string, f cmdFunc)                          { g_cmds[key] = f }
func _Rpc_cmd_tcp(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_cmd_http(req, ack) }
func _Rpc_cmd_http(req, ack *common.NetPack) {
	cmd := req.ReadString()

	args := strings.Split(cmd, " ")
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover HandleCmd\n%v: %s", r, debug.Stack())
		}
	}()
	if cmd, ok := g_cmds[args[0]]; ok {
		cmd(args)
		return
	}
}

// ------------------------------------------------------------
//! 命令行函数
func cmd_LogLv(args []string) {
	lv, _ := strconv.Atoi(args[1])
	gamelog.SetLevel(lv)
	fmt.Println("SetLogLv ", lv)
}
func cmd_gc(args []string) {
	runtime.GC()
	fmt.Println("GC finished")
}
func cmd_routine(args []string) {
	fmt.Println("Current number of goroutines: ", runtime.NumGoroutine())
}
func cmd_cpu(args []string) {
	fmt.Println(runtime.NumCPU(), " cpus and ", runtime.GOMAXPROCS(0), " in use")
}
func cmd_setcpu(args []string) {
	if conf.IsDebug {
		n, _ := strconv.Atoi(args[1])
		runtime.GOMAXPROCS(n)
		fmt.Println(runtime.NumCPU(), " cpus and ", runtime.GOMAXPROCS(0), " in use")
	}
}
