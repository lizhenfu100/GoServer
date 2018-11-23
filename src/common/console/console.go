package console

import (
	"bufio"
	"common"
	"conf"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"tcp"
)

type Func func(args []string)

var g_funcs = map[string]Func{
	"loglv":   cmd_LogLv,
	"gc":      cmd_gc,
	"routine": cmd_routine,
	"cpu":     cmd_cpu,
	"setcpu":  cmd_setcpu,
}

func Init() {
	tcp.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_cmd_tcp
	http.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_cmd_http
	//go _loop()
}
func _Rpc_cmd_tcp(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_cmd_http(req, ack) }
func _Rpc_cmd_http(req, ack *common.NetPack) {
	cmd := req.ReadString()
	HandleCmd(strings.Split(cmd, " "))
}
func _loop() {
	command := make([]byte, 1024)
	reader := bufio.NewReader(os.Stdin)
	for {
		command, _, _ = reader.ReadLine()
		args := strings.Split(string(command), " ")
		HandleCmd(args)
	}
}

func RegCmd(key string, cmd Func) { g_funcs[key] = cmd }

func HandleCmd(args []string) {
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover HandleCmd\n%v: %s", r, debug.Stack())
		}
	}()
	if cmd, ok := g_funcs[args[0]]; ok {
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
