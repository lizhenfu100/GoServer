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

var (
	_addr string
)

func init() {
	flag.StringVar(&_addr, "addr", "", "远端地址列表")
}

func main() {
	gamelog.InitLogger("SyncPatch")
	flag.Parse() //内部获取了所有参数：os.Args[1:]
	list := strings.Split(_addr, " ")
	for _, addr := range list {
		SyncServerPatch(fmt.Sprintf("http://%s/", addr))
	}
	time.Sleep(time.Hour)
}

// --------------------------------------------------------------------------
// 将本地文件上传至服务器
func SyncServerPatch(addr string) {
	http.CallRpc(addr, enum.Rpc_file_update_list, func(buf *common.NetPack) {
		buf.WriteString("")
	}, func(backBuf *common.NetPack) {
		if cnt := backBuf.ReadUInt16(); cnt > 0 {
			//服务器文件列表
			svrList := make(map[string]uint32, cnt)
			for i := uint16(0); i < cnt; i++ {
				filename := backBuf.ReadString()
				md5hash := backBuf.ReadUInt32()
				svrList[filename] = md5hash //记录远端文件名、MD5
			}
			backBuf.ReadString() //version

			//本地文件列表
			localList := make(map[string]uint32, cnt)
			names, _ := file.WalkDir("./", "")
			for _, v := range names {
				if strings.Index(v, "SyncPatch") == -1 {
					localList[v] = common.StringHash(file.CalcMd5(v))
				}
			}

			//比对本地、服务器，有新增或变更才上传
			for k, v1 := range localList {
				if v2, ok := svrList[k]; ok {
					if v1 != v2 { //变动文件
						http.UploadFile(addr+"upload_patch_file", k)
					}
				} else { //新增文件
					http.UploadFile(addr+"upload_patch_file", k)
				}
			}
		}
	})
}
