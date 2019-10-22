/***********************************************************************
* @ web请假小工具
* @ brief
	1、浏览器 url 默认是 "GET" 方法
		· 后台收到GET请求，返回相应html文件
		· html中写</form>"POST"

	2、html中引用其它文件(如js、css、jpeg)
		· 须由http.FileServer映射到正确目录
		· 默认下载
			· http.Handle("/", http.FileServer(http.Dir(kFileDirRoot)))
		· 自定义函数做下载控制
			· g_file_server = http.FileServer(http.Dir(kFileDirRoot))
			· http.HandleFunc("/", Http_download_file)
			· func Http_download_file(w http.ResponseWriter, r *http.Request) {
			·   //...
			·   g_file_server.ServeHTTP(w, r)
			·   //...
			· }

* @ author zhoumf
* @ date 2018-11-8
***********************************************************************/
package main

import (
	"common/file"
	"conf"
	"dbmgo"
	"flag"
	"gamelog"
	"netConfig"
	"netConfig/meta"
	"nets"
	"nets/http"
	http2 "nets/http/http"
)

func main() {
	ip, port := "", 0
	flag.StringVar(&ip, "ip", "192.168.1.111", "ip")
	flag.IntVar(&port, "port", 7702, "port")
	flag.Parse() //内部获取了所有参数：os.Args[1:]

	meta.G_Local = &meta.Meta{
		Module:   "AFK",
		SvrName:  "ChillyRoom_AFK",
		IP:       ip,
		OutIP:    ip,
		HttpPort: uint16(port),
	}

	//初始化日志系统
	gamelog.InitLogger("afk")
	http.InitClient(http2.Client)
	InitConf()

	file.TemplateDir(meta.G_Local, kFileDirRoot+"template/", kFileDirRoot, ".html")

	dbmgo.InitWithUser("", 27017, "other", conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	netConfig.RunNetSvr()
}
func InitConf() {
	file.LoadCsv("csv/conf_svr.csv", &conf.SvrCsv)

	nets.RegHttpHandler(map[string]nets.HttpHandle{
		"/afk":    Http_afk,
		"/del":    Http_del,
		"/count":  Http_count,
		"/my_afk": Http_my_afk,
	})
}
