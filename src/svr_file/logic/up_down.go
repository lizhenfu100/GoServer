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
	"conf"
	"fmt"
	"gamelog"
	"io"
	"net/http"
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
	file, handler, err := r.FormFile("file")
	if err != nil {
		gamelog.Error(err.Error())
		return
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)

	filename := "./net_file/upload/" + handler.Filename
	if err := os.MkdirAll(filepath.Dir(filename), 0777); err != nil {
		gamelog.Error(err.Error())
		return
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		gamelog.Error(err.Error())
		return
	}
	defer f.Close()

	io.Copy(f, file)
}
func Rpc_file_update_list(req, ack *common.NetPack) {
	version := req.ReadString()
	if strings.Compare(version, conf.SvrCsv.Version) < 0 {
		//下发 patch 目录下的文件列表
		names, _ := common.WalkDir("net_file/patch", "")
		ack.WriteUInt16(uint16(len(names)))
		for _, v := range names {
			ack.WriteString(strings.Trim(v, "net_file"))
		}
		ack.WriteString(conf.SvrCsv.Version)
	} else {
		ack.WriteUInt16(0)
	}
}
