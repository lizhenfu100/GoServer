package main

import (
	"common"
	"common/format"
	"common/std"
	"flag"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"io"
	"nets/http"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	var addrList, fileList string
	flag.StringVar(&addrList, "addr", "", "远端地址列表,空格隔开")
	flag.StringVar(&fileList, "file",
		"China/conf_svr.csv>csv/conf_svr.csv",
		"本地文件列表,空格隔开,可重定向成后台路径")
	flag.Parse()
	gamelog.InitLogger("UpdateCsv")

	addrList = format.MergeNearSpace(addrList)
	fileList = format.MergeNearSpace(fileList)

	//待上传的文件列表
	svrFile := make(map[string][]byte, 16)
	for _, str := range strings.Split(fileList, " ") {
		list := strings.Split(str, ">")
		fileLocal, fileSvr := "", "" //本地文件路径重定向为后台路径
		if len(list) == 1 {
			fileLocal = list[0]
			fileSvr = fileLocal
		} else {
			fileLocal = list[0]
			fileSvr = list[1]
		}
		if f, e := os.Open(fileLocal); e == nil {
			if fi, e := f.Stat(); e == nil {
				buf := make([]byte, fi.Size())
				if _, e = io.ReadFull(f, buf); e == nil {
					//fmt.Println("---------------1", fileSvr)
					svrFile[fileSvr] = buf
				}
			}
		}
	}
	addrs := strings.Split(addrList, " ")
	for _, addr := range _UpdateList(addrs) {
		UpdateCsv("http://"+addr, svrFile)
	}
	for _, addr := range addrs {
		ReloadCsv("http://"+addr, svrFile)
	}
	fmt.Println("\n...finish...")
	time.Sleep(time.Hour)
}

// ------------------------------------------------------------
func UpdateCsv(addr string, svrFile map[string][]byte) {
	http.CallRpc(addr, enum.Rpc_update_csv, func(buf *common.NetPack) {
		buf.WriteByte(byte(len(svrFile)))
		fmt.Println("\nstart update:", addr)
		for k, v := range svrFile {
			dir, name := path.Split(k)
			buf.WriteString(dir)
			buf.WriteString(name)
			buf.WriteLenBuf(v)
			fmt.Println("	", k, len(v))
		}
	}, func(recvBuf *common.NetPack) {
		fmt.Println("end:", recvBuf.ReadString())
	})
}
func ReloadCsv(addr string, svrFile map[string][]byte) {
	http.CallRpc(addr, enum.Rpc_reload_csv, func(buf *common.NetPack) {
		buf.WriteByte(byte(len(svrFile)))
		fmt.Println("\nstart reload:", addr)
		for k, _ := range svrFile {
			buf.WriteString(k)
			fmt.Println("	", k)
		}
	}, func(recvBuf *common.NetPack) {
		fmt.Println("end:", recvBuf.ReadByte())
	})
}

// ------------------------------------------------------------
// 每个均得Reload，同ip的选一个Update即可
func _UpdateList(addrs []string) (ret []string) {
	var ips std.Strings
	for _, addr := range addrs {
		if idx := strings.Index(addr, ":"); idx >= 0 {
			ip := addr[:idx]
			if ips.Index(ip) < 0 {
				ips.Add(ip)
				ret = append(ret, addr)
			}
		}
	}
	return
}
