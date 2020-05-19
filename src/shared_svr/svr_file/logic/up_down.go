/***********************************************************************
* @ 客户端热更新
* @ brief
    1、客户端启动，先连svr_file，上报自己的版本号，同后台比对
    2、不一致则将后台 net_file/patch 目录中的(文件名, md5)下发
    3、客户端据收到的文件列表，逐个同本地比对，不一致的才向后台下载

* @ 下载
	· HttpUpDn.cs http下载，可cdn加速

* @ Notice
	· 此类全局性资源，为避免热更引发竞态卡顿，应保持旧资源有效性，新资源生成后交换引用

* @ author zhoumf
* @ date 2017-8-23
***********************************************************************/
package logic

import (
	"common"
	"common/console"
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
)

var (
	g_file_md5 sync.Map //<fileName, md5Hash>
)

func init() {
	names, _ := file.WalkDir(kFileDirPatch, "")
	for _, v := range names {
		g_file_md5.Store(v, file.CalcMd5(v))
	}
	console.RegCmd("clog", Client_debug_log)
}
func Http_upload_patch_file(w http.ResponseWriter, r *http.Request) {
	if name := _upload_file(r, kFileDirPatch); name != "" {
		g_file_md5.Store(name, file.CalcMd5(name))
	}
}
func Http_upload_player_file(w http.ResponseWriter, r *http.Request) {
	_upload_file(r, kFileDirPlayer)
}
func _upload_file(r *http.Request, baseDir string) string {
	upfile, h, err := r.FormFile("file")
	if err != nil {
		gamelog.Error(err.Error())
		return ""
	}
	defer upfile.Close()

	fullname := baseDir + h.Filename
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
	os.Rename(fullname+"_2", fullname)
	return fullname
}

func Rpc_file_update_list(req, ack *common.NetPack, _ common.Conn) {
	version := req.ReadString()
	destFolder := req.ReadString()

	posInBuf, count := ack.Size(), uint32(0)
	ack.WriteUInt32(count)

	if common.IsMatchVersion(meta.G_Local.Version, version) {
		//下发patch目录下的文件列表
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
		ack.SetUInt32(posInBuf, count)
	}
}
func Rpc_file_delete(req, ack *common.NetPack, _ common.Conn) {
	for cnt, i := req.ReadInt(), 0; i < cnt; i++ {
		name := kFileDirPatch + req.ReadString()
		g_file_md5.Delete(name)
		os.Remove(name)
	}
}
