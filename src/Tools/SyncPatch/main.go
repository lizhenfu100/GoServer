package main

import (
	"common"
	"common/file"
	"flag"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"http"
	"strings"
	"time"
)

func main() {
	var sleep bool
	var addrList, dirList string
	flag.StringVar(&addrList, "addr", "", "远端地址列表")
	flag.StringVar(&dirList, "dir", "", "本地目录列表")
	flag.BoolVar(&sleep, "sleep", true, "")
	flag.Parse() //内部获取了所有参数：os.Args[1:]
	gamelog.InitLogger("SyncPatch")

	//本地文件列表
	localMap := make(map[string]uint32, 32)
	for _, dir := range strings.Split(dirList, " ") {
		if dir != "" {
			readDir(dir, localMap)
		}
	}
	for _, addr := range strings.Split(addrList, " ") {
		if addr != "" {
			SyncServerPatch(fmt.Sprintf("http://%s", addr), localMap)
		}
	}
	if fmt.Println("\n...finish..."); sleep {
		time.Sleep(time.Hour)
	}
}

// --------------------------------------------------------------------------
// 将SyncPatch.exe所在目录的文件上传至服务器
func SyncServerPatch(addr string, localMap map[string]uint32) {
	http.CallRpc(addr, enum.Rpc_file_update_list, func(buf *common.NetPack) {
		buf.WriteString("") //version
		buf.WriteString("")
	}, func(backBuf *common.NetPack) {
		svrMap := make(map[string]uint32, 32)
		if cnt := backBuf.ReadUInt32(); cnt > 0 {
			//服务器文件列表
			for i := uint32(0); i < cnt; i++ {
				name := backBuf.ReadString()
				md5hash := backBuf.ReadUInt32()
				svrMap[name] = md5hash //记录远端文件名、MD5
			}
		}

		//本次无需同步的文件列表，用于显示
		notSyncList := make([]string, 0, 32)

		//比对本地、服务器，有新增或变更才上传
		fmt.Println("\nSync File: ")
		for k, v1 := range localMap {
			if v2, ok := svrMap[k]; ok {
				if v1 != v2 { //变动文件
					if http.UploadFile(addr+"/upload_patch_file", k) == nil {
						fmt.Println("    ", k)
					}
				} else {
					notSyncList = append(notSyncList, k)
				}
			} else { //新增文件
				if http.UploadFile(addr+"/upload_patch_file", k) == nil {
					fmt.Println("    ", k)
				}
			}
		}

		//打印本次无需同步的文件
		fmt.Println("\nNot Sync File: ")
		for _, v := range notSyncList {
			fmt.Println("    ", v)
		}
	})
}
func readDir(dir string, localMap map[string]uint32) {
	names, _ := file.WalkDir(dir, "")
	for _, v := range names {
		v = strings.TrimPrefix(v, "./")
		localMap[v] = common.StringHash(file.CalcMd5(v))
	}
}
