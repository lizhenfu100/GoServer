/***********************************************************************
* @ 客户端热更新
* @ brief
    1、客户端启动，先连svr_file，上报自己的版本号，同后台比对

    2、若小于后台的，将后台 patch 目录中的文件名打包下发

    3、客户端据收到的文件名列表，逐个【增量更新】

* @ Notice
	1、HttpDownload 是通过文件 seek 定位的，内容减少的变动，无法检测到

	2、有必要的话，可以加个 cover.txt 之类的，里头的问题无脑覆盖更新

* @ author zhoumf
* @ date 2017-8-23
***********************************************************************/
package logic

import (
	"common"
	"conf"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	// g_file_patch = http.StripPrefix("/download/", http.FileServer(http.Dir("patch")))
	g_file_patch  = http.FileServer(http.Dir("patch"))
	g_file_upload = http.FileServer(http.Dir("upload"))
)

func Handle_File_Download(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method, "url down path: "+r.URL.Path)
	if common.IsExist("patch" + r.URL.Path) {
		g_file_patch.ServeHTTP(w, r)
	} else if common.IsExist("upload" + r.URL.Path) {
		g_file_upload.ServeHTTP(w, r)
	}
}
func Handle_File_Upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method, "url up path: "+r.URL.Path)
	r.ParseMultipartForm(1024 * 1024)
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)

	filename := "./upload/" + handler.Filename
	if err := os.MkdirAll(filepath.Dir(filename), 0777); err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	io.Copy(f, file)
}
func Rpc_Update_File_List(req, ack *common.NetPack) {
	version := req.ReadString()
	if strings.Compare(version, conf.SvrCfg.Version) < 0 {
		//下发 patch 目录下的文件列表
		names, _ := common.WalkDir("patch", "")
		ack.WriteUInt16(uint16(len(names)))
		for _, v := range names {
			ack.WriteString(strings.Trim(v, "patch"))
		}
		ack.WriteString(conf.SvrCfg.Version)
	} else {
		ack.WriteUInt16(0)
	}
}
