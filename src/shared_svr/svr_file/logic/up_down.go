/***********************************************************************
* @ 客户端热更新
* @ brief
    1、客户端启动，先连svr_file，上报自己的版本号，同后台比对

    2、若小于后台的，将后台 net_file/patch 目录中的文件名打包下发

    3、客户端据收到的文件名列表，逐个【增量更新】

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
	"netConfig"
	"os"
	"path/filepath"
	"sync"
)

const (
	Http_File_Dir   = "net_file"
	Player_File_Dir = "net_file/upload/"
	Patch_File_Dir  = "net_file/patch/"
)

var (
	G_file_md5 sync.Map
)

func init() {
	http.Handle("/", http.FileServer(http.Dir(Http_File_Dir)))

	names, _ := file.WalkDir(Patch_File_Dir, "")
	for _, v := range names {
		md5str := file.CalcMd5(v)
		G_file_md5.Store(v, common.StringHash(md5str))
	}
}

func Http_upload_player_file(w http.ResponseWriter, r *http.Request) {
	_upload_file(w, r, Player_File_Dir)
}
func Http_upload_patch_file(w http.ResponseWriter, r *http.Request) {
	name := _upload_file(w, r, Patch_File_Dir)
	md5str := file.CalcMd5(name)
	G_file_md5.Store(name, common.StringHash(md5str))
}
func _upload_file(w http.ResponseWriter, r *http.Request, baseDir string) string {
	gamelog.Debug("%s url up path: %s", r.Method, r.URL.Path)
	r.ParseMultipartForm(1024 * 1024)
	upfile, handler, err := r.FormFile("file")
	if err != nil {
		gamelog.Error(err.Error())
		return ""
	}
	defer upfile.Close()

	fullname := baseDir + handler.Filename
	dir, name := filepath.Dir(fullname), filepath.Base(fullname)

	if f, err := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC); err == nil {
		io.Copy(f, upfile)
		f.Close()
	} else {
		gamelog.Error(err.Error())
	}
	return fullname
}

func Rpc_file_update_list(req, ack *common.NetPack) {
	version := req.ReadString()
	if version == netConfig.G_Local_Meta.Version {
		ack.WriteUInt16(0)
	} else {
		//下发 patch 目录下的文件列表
		names, _ := file.WalkDir(Patch_File_Dir, "")
		ack.WriteUInt16(uint16(len(names)))
		for _, v := range names {
			ack.WriteString(v[len(Patch_File_Dir):])
			vv, _ := G_file_md5.Load(v)
			ack.WriteUInt32(vv.(uint32))
		}
		ack.WriteString(netConfig.G_Local_Meta.Version)
	}
}
