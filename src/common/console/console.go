package console

import (
	"common"
	"common/file"
	"common/wechat"
	"conf"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"http"
	"math/rand"
	"netConfig/meta"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"tcp"
	"time"
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
	rand.Seed(time.Now().Unix())
	tcp.G_HandleFunc[enum.Rpc_log] = _Rpc_log1
	http.G_HandleFunc[enum.Rpc_log] = _Rpc_log2
	tcp.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_gm_cmd1
	http.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_gm_cmd2
	tcp.G_HandleFunc[enum.Rpc_meta_list] = _Rpc_meta_list1
	http.G_HandleFunc[enum.Rpc_meta_list] = _Rpc_meta_list2
	tcp.G_HandleFunc[enum.Rpc_update_csv] = _Rpc_update_csv1
	http.G_HandleFunc[enum.Rpc_update_csv] = _Rpc_update_csv2
	go sigTerm() //监控进程终止信号

	wechat.Init( //启动微信通知
		conf.SvrCsv.WechatCorpId,
		conf.SvrCsv.WechatSecret,
		conf.SvrCsv.WechatTouser,
		conf.SvrCsv.WechatAgentId)
}

func _Rpc_log1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_log2(req, ack) }
func _Rpc_log2(req, ack *common.NetPack) {
	log := req.ReadString()
	uuid := req.ReadString()
	version := req.ReadString()
	platform := req.ReadString()
	gamelog.Info("%s\nUUID: %s version: %s platform: %s", log, uuid, version, platform)
}

func _Rpc_meta_list1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_meta_list2(req, ack) }
func _Rpc_meta_list2(req, ack *common.NetPack) {
	module := req.ReadString() //game、save、file...
	version := req.ReadString()

	ids := meta.GetModuleIDs(module, version)
	sort.Ints(ids)
	ack.WriteByte(byte(len(ids)))
	for _, id := range ids {
		p := meta.GetMeta(module, id)
		ack.WriteInt(p.SvrID)
		ack.WriteString(p.OutIP)
		ack.WriteUInt16(p.Port())
		ack.WriteString(p.SvrName)
	}
}

func _Rpc_update_csv1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_update_csv2(req, ack) }
func _Rpc_update_csv2(req, ack *common.NetPack) {
	for cnt, i := req.ReadByte(), byte(0); i < cnt; i++ {
		dir := req.ReadString()
		name := req.ReadString()
		data := req.ReadLenBuf()
		if fd, e := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC); e == nil {
			_, e = fd.Write(data)
			fd.Close()
			if e == nil {
				ack.WriteString("ok")
			} else {
				ack.WriteString(e.Error())
			}
		} else {
			ack.WriteString(e.Error())
		}
	}
}

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
