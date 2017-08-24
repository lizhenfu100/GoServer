package common

import (
	"bufio"
	"fmt"
	"gamelog"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type CommandHandler func(args []string) bool

var G_HandlerMap = map[string]CommandHandler{
	"setloglevel": HandCmd_SetLogLevel,
}

func StartConsole() {
	go consoleroutine()
}
func consoleroutine() {
	command := make([]byte, 1024)
	reader := bufio.NewReader(os.Stdin)
	for {
		command, _, _ = reader.ReadLine()
		args := strings.Split(string(command), " ")

		if cmdhandler, ok := G_HandlerMap[args[0]]; ok {
			cmdhandler(args)
			continue
		}

		switch string(args[0]) {
		case "cpus":
			fmt.Println(runtime.NumCPU(), " cpus and ", runtime.GOMAXPROCS(0), " in use")

		case "routines":
			fmt.Println("Current number of goroutines: ", runtime.NumGoroutine())

		case "setcpus":
			n, _ := strconv.Atoi(args[1])
			runtime.GOMAXPROCS(n)
			fmt.Println(runtime.NumCPU(), " cpus and ", runtime.GOMAXPROCS(0), " in use")

		case "startgc":
			runtime.GC()
			fmt.Println("gc finished")
		default:
			fmt.Println("Command error, try again.")
		}
	}
}
func RegConsoleCmd(cmd string, mh CommandHandler) {
	G_HandlerMap[cmd] = mh
}

//////////////////////////////////////////////////////////////////////
//! 命令行函数
func HandCmd_SetLogLevel(args []string) bool {
	if len(args) < 2 {
		gamelog.Error("Lack of param")
		return false
	}
	level, err := strconv.Atoi(args[1])
	if err != nil {
		gamelog.Error("HandCmd_SetLogLevel => Invalid param:%s", args[1])
		return false
	}
	gamelog.SetLevel(level)
	return true
}
