/***********************************************************************
* @ GM系统
* @ brief
	1、先从center拉取所有login地址，再拉login下所有game地址

	2、填充模板，生成真正的HTML文件，方便查看

	3、密码每周一零点更新（记log），输入错误，将密码发往对应邮箱

* @ author zhoumf
* @ date 2019-2-20
***********************************************************************/
package main

import (
	"common/file"
	"conf"
	"flag"
	"gamelog"
	"net/http"
	"netConfig"
	"netConfig/meta"
	"netConfig/register"
	mhttp "nets/http"
	http2 "nets/http/http"
	"strings"
)

const (
	kFileDirRoot = "html/GM/"
	kNeedPasswd  = false
)

func main() {
	ip, port, _getaddrs := "", 0, false
	flag.StringVar(&ip, "ip", "192.168.1.111", "ip")
	flag.IntVar(&port, "port", 7701, "port")
	flag.BoolVar(&_getaddrs, "getaddrs", false, "")
	flag.Parse() //内部获取了所有参数：os.Args[1:]

	meta.G_Local = &meta.Meta{
		Module:   "GM",
		SvrName:  "ChillyRoom_GM",
		IP:       ip,
		OutIP:    ip,
		HttpPort: uint16(port),
	}
	g_common.LocalAddr = mhttp.Addr(ip, uint16(port))

	//初始化日志系统
	gamelog.InitLogger("gm")
	mhttp.InitClient(http2.Client)
	InitConf()

	//地址信息
	for i := 0; i < len(g_list); i++ {
		p := &g_list[i]
		p._TCommon = g_common
		if !p.LoadAddrs() || _getaddrs {
			p.GetAddrs()
		}
		UpdateHtmls("game", "game."+p.GameName, p) //game中的页面往各游戏都导一份
		UpdateHtmls("game."+p.GameName, "game."+p.GameName, p)
	}
	UpdateHtmls("account/", "account/", g_common)
	UpdateHtml("index", "index", g_common)
	UpdateHtml("passwd", "passwd", g_common)
	UpdateHtml("relay_gm_cmd", "relay_gm_cmd", g_common)

	if kNeedPasswd {
		go UpdatePasswd()
	}
	netConfig.RunNetSvr()
}
func InitConf() {
	file.LoadCsv("csv/conf_svr.csv", &conf.SvrCsv)

	register.RegHttpHandler(map[string]register.HttpHandle{
		"/query_account_login_addr": Http_query_account_login_addr,
		"/reset_password":           Http_reset_password,
		"/check_passwd":             Http_check_passwd,
		"/backup_conf":              Http_relay_to_save,
		"/backup_auto":              Http_relay_to_save,
		"/backup_force":             Http_relay_to_save,
		"/relay_gm_cmd":             Http_relay_gm_cmd,
	})
	g_file_server = http.FileServer(http.Dir(kFileDirRoot))
	http.HandleFunc("/", Http_download_file)
}

var g_file_server http.Handler

func Http_download_file(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("download path: " + r.URL.Path)
	if kNeedPasswd && r.URL.Path == "/" {
		r.URL.Path = "/passwd.html"
	}
	if strings.HasSuffix(r.URL.Path, "app.js") {
		r.URL.Path = "/app.js"
	}
	if strings.HasSuffix(r.URL.Path, "main.css") {
		r.URL.Path = "/main.css"
	}
	g_file_server.ServeHTTP(w, r)
}
