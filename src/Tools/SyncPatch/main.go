package main

import (
	"common"
	"common/file"
	"common/format"
	"common/tool/cdn"
	"flag"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"nets/http"
	"strings"
	"time"
)

const (
	kCDNAddr = "http://111.230.177.63:7071" //官网
	kCDNUrl  = "http://www.chillyroom.com/game/"
)

func main() {
	var sleep int
	var addrList, dirList string
	flag.StringVar(&addrList, "addr", "", "远端地址列表,空格隔开")
	flag.StringVar(&dirList, "dir", "", "本地目录列表,空格隔开")
	flag.IntVar(&sleep, "sleep", 1, "")
	flag.Parse()
	gamelog.InitLogger("SyncPatch")

	addrList = format.MergeNearSpace(addrList)
	dirList = format.MergeNearSpace(dirList)

	//本地文件列表
	localMap := make(map[string]uint32, 32)
	localDirs := strings.Split(dirList, " ")
	for _, dir := range localDirs {
		readDir(dir, localMap)
	}
	for _, addr := range strings.Split(addrList, " ") {
		SyncServerPatch("http://"+addr, localMap, localDirs)
	}
	if fmt.Println("\n...finish..."); sleep > 0 {
		time.Sleep(time.Hour)
	}
}
func InDirs(name string, dirs []string) bool {
	for _, v := range dirs {
		if strings.HasPrefix(name, v) {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------
func SyncServerPatch(addr string, localMap map[string]uint32, localDirs []string) {
	http.CallRpc(addr, enum.Rpc_file_update_list, func(buf *common.NetPack) {
		buf.WriteString("") //version
		buf.WriteString("")
	}, func(backBuf *common.NetPack) {
		svrMap := make(map[string]uint32, 32)
		if cnt := backBuf.ReadUInt32(); cnt > 0 {
			for i := uint32(0); i < cnt; i++ {
				name := backBuf.ReadString()
				md5hash := backBuf.ReadUInt32()

				//记录同路径下，远端文件名、MD5
				if InDirs(name, localDirs) {
					svrMap[name] = md5hash
				}
			}
		}

		notSyncList := make([]string, 0, 32) //无需同步的，用于显示
		cdnUrls := make([]string, 0, 32)     //同步过的，用于刷新cdn

		//比对本地、服务器，有新增或变更才上传
		fmt.Println("\nSync File: ")
		for k, v1 := range localMap {
			if v2, ok := svrMap[k]; ok {
				if v1 != v2 { //变动文件
					if http.UploadFile(addr+"/upload_patch_file", k) == nil {
						cdnUrls = append(cdnUrls, kCDNUrl+k)
						fmt.Println("    ", k)
					}
				} else {
					notSyncList = append(notSyncList, k)
				}
				delete(svrMap, k) //删除本地有的，剩余即服务器上待删除的
			} else { //新增文件
				if http.UploadFile(addr+"/upload_patch_file", k) == nil {
					cdnUrls = append(cdnUrls, kCDNUrl+k)
					fmt.Println("    ", k)
				}
			}
		}
		//删除服务器旧文件
		if len(svrMap) > 0 {
			fmt.Println("\nDelete Old File: ")
			http.CallRpc(addr, enum.Rpc_file_delete, func(buf *common.NetPack) {
				buf.WriteInt(len(svrMap))
				for k := range svrMap {
					fmt.Println("    ", k)
					buf.WriteString(k)
					cdnUrls = append(cdnUrls, kCDNUrl+k)
				}
			}, nil)
		}
		//刷新官网文件的cdn
		if addr == kCDNAddr && len(cdnUrls) > 0 {
			fmt.Println("\nRefresh CDN: ")
			fmt.Println("    ", cdn.RefreshUrl(cdnUrls))
		}
		//打印无需同步的文件
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
		localMap[v] = file.CalcMd5(v)
	}
}
