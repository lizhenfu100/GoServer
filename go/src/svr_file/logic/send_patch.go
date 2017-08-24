/***********************************************************************
* @ 客户端热更新
* @ brief
    1、客户端启动，先连svr_file，上报自己的版本号，同后台比对

    2、若小于后台的，将后台 patch 目录中的文件全部下发

* @ author zhoumf
* @ date 2017-8-23
***********************************************************************/
package logic

import (
	"common"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	G_Version = "1.0.1"
)

var (
	// g_file_handler = http.StripPrefix("/download/", http.FileServer(http.Dir("patch")))
	g_file_handler = http.FileServer(http.Dir("patch"))
)

func Handle_File_Download(w http.ResponseWriter, req *http.Request) {
	fmt.Println("url path:" + req.URL.Path)
	g_file_handler.ServeHTTP(w, req)
}
func Handle_File_Upload(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
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
}
func Rpc_Update_File_List(req, ack *common.NetPack) {
	version := req.ReadString()
	if strings.Compare(version, G_Version) < 0 {
		//下发 patch 目录下的文件列表
		names, _ := common.WalkDir("patch", "")
		ack.WriteUInt16(uint16(len(names)))
		for _, v := range names {
			ack.WriteString(strings.Trim(v, "patch"))
		}
		ack.WriteString(G_Version)
	} else {
		ack.WriteUInt16(0)
	}
}
