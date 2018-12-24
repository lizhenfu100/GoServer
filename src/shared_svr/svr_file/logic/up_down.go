/***********************************************************************
* @ 客户端热更新
* @ brief
    1、客户端启动，先连svr_file，上报自己的版本号，同后台比对
    2、不一致则将后台 net_file/patch 目录中的(文件名, md5)下发
    3、客户端据收到的文件列表，逐个同本地比对，不一致的才向后台下载

* @ Notice
	1、HttpDownload 是通过文件 seek 定位的，内容减少的变动，无法检测到
	2、有必要的话，可以加个 cover.txt 之类的，里头的文件无脑覆盖更新

* @ author zhoumf
* @ date 2017-8-23
***********************************************************************/
package logic

import (
	"common"
	"common/file"
	"gamelog"
	"io"
	"net/http"
	"netConfig/meta"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	kFileDirRoot   = "net_file"
	kFileDirPlayer = "net_file/upload/"
	kFileDirPatch  = "net_file/patch/"
	kMaxSizeUpload = 1024 * 1024
)

var (
	g_file_md5    sync.Map //<fileName, md5Hash>
	g_file_server http.Handler
	g_file_mutex  sync.RWMutex
)

func init() {
	//http.Handle("/", http.FileServer(http.Dir(kFileDirRoot)))
	g_file_server = http.FileServer(http.Dir(kFileDirRoot))
	http.HandleFunc("/", Http_download_file)

	names, _ := file.WalkDir(kFileDirPatch, "")
	for _, v := range names {
		md5str := file.CalcMd5(v)
		g_file_md5.Store(v, common.StringHash(md5str))
	}
}
func Http_download_file(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("download path: %s", r.URL.Path)
	g_file_mutex.RLock()
	g_file_server.ServeHTTP(w, r) //读patch目录下的文件
	g_file_mutex.RUnlock()
}
func Http_upload_patch_file(w http.ResponseWriter, r *http.Request) {
	g_file_mutex.Lock()
	name := _upload_file(w, r, kFileDirPatch) //写patch目录下的文件
	g_file_mutex.Unlock()
	md5str := file.CalcMd5(name)
	g_file_md5.Store(name, common.StringHash(md5str))
}
func Http_upload_player_file(w http.ResponseWriter, r *http.Request) {
	_upload_file(w, r, kFileDirPlayer)
}
func _upload_file(w http.ResponseWriter, r *http.Request, baseDir string) string {
	r.ParseMultipartForm(kMaxSizeUpload)
	upfile, handler, err := r.FormFile("file")
	if err != nil {
		gamelog.Error(err.Error())
		return ""
	}
	defer upfile.Close()

	fullname := baseDir + handler.Filename
	gamelog.Debug("Path:%s  Name:%s", r.URL.Path, fullname)

	dir, name := filepath.Split(fullname)
	if f, err := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC); err == nil {
		io.Copy(f, upfile)
		f.Close()
	} else {
		gamelog.Error(err.Error())
		return ""
	}
	return fullname
}

func Rpc_file_update_list(req, ack *common.NetPack) {
	version := req.ReadString()
	destFolder := req.ReadString()
	if meta.G_Local.IsMatchVersion(version) {
		//下发patch目录下的文件列表
		posInBuf, count := ack.BodySize(), uint32(0)
		ack.WriteUInt32(count)
		g_file_md5.Range(func(k, v interface{}) bool {
			fileName := k.(string)[len(kFileDirPatch):] //patch后的文件路径
			if strings.HasPrefix(fileName, destFolder) {
				ack.WriteString(fileName)
				ack.WriteUInt32(v.(uint32))
				count++
			}
			return true
		})
		ack.SetPos(posInBuf, count)
	} else {
		ack.WriteUInt16(0)
	}
}
