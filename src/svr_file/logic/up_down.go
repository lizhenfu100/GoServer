/***********************************************************************
* @ 客户端热更新
* @ brief
    1、客户端启动，先连svr_file，上报自己的版本号，同后台比对

    2、若小于后台的，将后台 net_file/patch 目录中的文件名打包下发

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
	"common/file"
	"fmt"
	"gamelog"
	"io"
	"net/http"
	"netConfig"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	http.Handle("/", http.FileServer(http.Dir("net_file")))
}

func Http_file_upload(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("%s url up path: %s", r.Method, r.URL.Path)
	r.ParseMultipartForm(1024 * 1024)
	upfile, handler, err := r.FormFile("file")
	if err != nil {
		gamelog.Error(err.Error())
		return
	}
	defer upfile.Close()
	fmt.Fprintf(w, "%v", handler.Header)

	path := "./net_file/upload/" + handler.Filename
	dir, name := filepath.Dir(path), filepath.Base(path)

	f, err := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		gamelog.Error(err.Error())
		return
	}
	defer f.Close()

	io.Copy(f, upfile)
}
func Rpc_file_update_list(req, ack *common.NetPack) {
	version := req.ReadString()
	//TODO：可动态更改节点版本号
	if strings.Compare(version, netConfig.G_Local_Meta.Version) < 0 {
		//下发 patch 目录下的文件列表
		names, _ := file.WalkDir("net_file/patch", "")
		ack.WriteUInt16(uint16(len(names)))
		for _, v := range names {
			ack.WriteString(strings.Trim(v, "net_file"))
		}
		ack.WriteString(netConfig.G_Local_Meta.Version)
	} else {
		ack.WriteUInt16(0)
	}
}
