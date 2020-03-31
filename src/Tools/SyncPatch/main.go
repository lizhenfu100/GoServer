package main

import (
	"common"
	"common/file"
	"common/format"
	"common/std"
	"common/tool/cdn"
	"flag"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"nets/http"
	_ "nets/http/http"
	"strings"
	"time"
)

const (
	kCDNAddr = "http://111.230.177.63:7071" //官网
	kCDNUrl  = "http://www.chillyroom.com/game/"
	// Notice：官网资源，须加项目名称的父目录，以防项目间的文件同名
)

func main() {
	var sleep int
	var addrList, dirList, gameName string
	flag.StringVar(&addrList, "addr", "", "远端地址列表,空格隔开")
	flag.StringVar(&dirList, "dir", "", "本地目录列表,空格隔开")
	flag.StringVar(&gameName, "game", "", "游戏名称")
	flag.IntVar(&sleep, "sleep", 1, "")
	flag.Parse()
	gamelog.InitLogger("SyncPatch")

	addrList = format.MergeNearSpace(addrList)
	dirList = format.MergeNearSpace(dirList)
	if addrList == "" || dirList == "" {
		fmt.Println("Param nil, run .bat")
	}

	//本地文件列表
	localMap := make(map[string]uint32, 32)
	localDirs := strings.Split(dirList, " ")
	for _, dir := range localDirs {
		readDir(dir, localMap)
	}
	for _, addr := range strings.Split(addrList, " ") {
		SyncServerPatch("http://"+addr, gameName, localMap, localDirs)
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
func SyncServerPatch(addr, gameName string, localMap map[string]uint32, localDirs []string) {
	kIsCdn := addr == kCDNAddr //是否官网节点
	http.CallRpc(addr, enum.Rpc_file_update_list, func(buf *common.NetPack) {
		buf.WriteString("") //version
		if kIsCdn {
			buf.WriteString(gameName) //destDir
		} else {
			buf.WriteString("") //destDir
		}
	}, func(backBuf *common.NetPack) {
		svrMap := make(map[string]uint32, 32)
		for cnt, i := backBuf.ReadUInt32(), uint32(0); i < cnt; i++ {
			name := backBuf.ReadString()
			md5hash := backBuf.ReadUInt32()
			//记录同路径下，远端文件名、MD5
			if InDirs(name, localDirs) {
				svrMap[name] = md5hash
			}
		}

		var notSyncList std.Strings //无需同步的，用于显示
		var cdnUrls std.Strings     //同步过的，用于刷新cdn

		//比对本地、服务器，有新增或变更才上传
		fmt.Println("\nSync File: ")
		for k, v1 := range localMap {
			if kIsCdn {
				k = gameName + "/" + k //官网资源，追加游戏名父目录
			}
			if v2, ok := svrMap[k]; ok {
				if v1 != v2 { //变动文件
					if e := http.UploadFile(addr+"/upload_patch_file", k); e == nil {
						cdnUrls.Add(kCDNUrl + k)
						fmt.Println("    ", k)
					} else {
						fmt.Println("upload fail: ", e.Error())
						return
					}
				} else {
					notSyncList.Add(k)
				}
				delete(svrMap, k) //删除本地有的，剩余即服务器上要删的
			} else { //新增文件
				if e := http.UploadFile(addr+"/upload_patch_file", k); e == nil {
					cdnUrls.Add(kCDNUrl + k)
					fmt.Println("    ", k)
				} else {
					fmt.Println("upload fail: ", e.Error())
					return
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
					cdnUrls.Add(kCDNUrl + k)
				}
			}, nil)
		}
		//刷新官网文件的cdn
		if kIsCdn && len(cdnUrls) > 0 {
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
