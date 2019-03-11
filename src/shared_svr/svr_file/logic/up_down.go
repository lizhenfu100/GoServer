/***********************************************************************
* @ 客户端热更新
* @ brief
    1、客户端启动，先连svr_file，上报自己的版本号，同后台比对
    2、不一致则将后台 net_file/patch 目录中的(文件名, md5)下发
    3、客户端据收到的文件列表，逐个同本地比对，不一致的才向后台下载

* @ Notice
	1、此类全局性资源，应使用新旧双备份策略，避免热更时引发竞态卡顿

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
	kFileDirRoot   = "net_file/"
	kFileDirPlayer = kFileDirRoot + "upload/"
	kFileDirPatch  = kFileDirRoot + "patch/"
	kMaxUploadSize = 1024 * 1024 * 20
)

var (
	g_file_md5    sync.Map //<fileName, md5Hash>
	g_file_server http.Handler
)

func init() {
	//http.Handle("/", http.FileServer(http.Dir(kFileDirRoot)))
	//带url路由的文件服务
	//http.Handle("/chillyroom_res/", http.StripPrefix("/chillyroom_res/", http.FileServer(http.Dir(kFileDirRoot))))
	//g_file_server = http.StripPrefix("/chillyroom_res/", http.FileServer(http.Dir(kFileDirRoot))）
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
	m := GetRWMutex(r.URL.Path)
	m.RLock()
	g_file_server.ServeHTTP(w, r)
	m.RUnlock()
}
func Http_upload_patch_file(w http.ResponseWriter, r *http.Request) {
	if name := _upload_file(w, r, kFileDirPatch); name != "" {
		g_file_md5.Store(name, common.StringHash(file.CalcMd5(name)))
	}
}
func Http_upload_player_file(w http.ResponseWriter, r *http.Request) {
	_upload_file(w, r, kFileDirPlayer)
}
func _upload_file(w http.ResponseWriter, r *http.Request, baseDir string) string {
	r.Body = http.MaxBytesReader(w, r.Body, kMaxUploadSize)
	if err := r.ParseMultipartForm(kMaxUploadSize); err != nil {
		gamelog.Error(err.Error())
		return ""
	}

	upfile, handler, err := r.FormFile("file")
	if err != nil {
		gamelog.Error(err.Error())
		return ""
	}
	defer upfile.Close()

	fullname := baseDir + handler.Filename
	gamelog.Debug("Path:%s  Name:%s", r.URL.Path, fullname)

	// 创建临时文件，避免直接写原文件带来的竞态
	dir, name := filepath.Split(fullname)
	if f, err := file.CreateFile(dir, name+"_2", os.O_WRONLY|os.O_TRUNC); err == nil {
		io.Copy(f, upfile)
		f.Close()
	} else {
		gamelog.Error(err.Error())
		return ""
	}
	// 加锁，用临时文件替代旧文件
	m := GetRWMutex(strings.TrimPrefix(fullname, kFileDirRoot))
	m.Lock()
	os.Rename(fullname+"_2", fullname)
	m.Unlock()
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
			name := strings.TrimPrefix(k.(string), kFileDirPatch) //patch后的文件路径
			//gamelog.Debug("---- svr file: %s %d", name, v.(uint32))
			if strings.HasPrefix(name, destFolder) {
				ack.WriteString(name)
				ack.WriteUInt32(v.(uint32))
				count++
			}
			return true
		})
		ack.SetPos(posInBuf, count)
	} else {
		ack.WriteUInt32(0)
	}
}

// ------------------------------------------------------------
// 细粒度的文件读写锁
var (
	_mutex_    sync.Mutex
	_rw_mutexs = make(map[string]*sync.RWMutex)
)

func GetRWMutex(name string) (ret *sync.RWMutex) {
	_mutex_.Lock()
	if v, ok := _rw_mutexs[name]; ok {
		ret = v
	} else {
		ret = new(sync.RWMutex)
		_rw_mutexs[name] = ret
	}
	_mutex_.Unlock()
	return
}
