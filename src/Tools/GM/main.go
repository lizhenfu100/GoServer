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
	"common/console"
	"common/file"
	"conf"
	"flag"
	"gamelog"
	"net/http"
	"netConfig"
	"netConfig/meta"
	"nets"
	mhttp "nets/http"
	http2 "nets/http/http"
	"strings"
)

const (
	kFileDirRoot = "html/GM/"
	kTemplateDir = "html/GM/template/"
)

func main() {
	ip, port, _getaddrs := "", 0, false
	flag.StringVar(&ip, "ip", "192.168.1.210", "ip")
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
	for k, v := range g_map {
		v.TCommon = g_common
		if !v.LoadAddrs() || _getaddrs {
			v.GetAddrs()
		}
		UpdateHtmls("game", "game."+v.GameName, &v) //game中的页面往各游戏都导一份
		UpdateHtmls("game."+v.GameName, "game."+v.GameName, &v)
		g_map[k] = v
	}
	UpdateHtmls("account/", "account/", g_common)
	UpdateHtml("index", "index", g_common)
	UpdateHtml("relay_gm_cmd", "relay_gm_cmd", g_common)

	netConfig.RunNetSvr()
}
func InitConf() {
	file.LoadCsv("csv/conf_svr.csv", &conf.SvrCsv)
	console.Init()

	nets.RegHttpHandler(map[string]nets.HttpHandle{
		"/reset_password":     Http_reset_password,
		"/bind_info_force":    Http_bind_info_force,
		"/backup_conf":        Http_relay_to_save,
		"/backup_auto":        Http_relay_to_save,
		"/backup_force":       Http_relay_to_save,
		"/relay_gm_cmd":       Http_relay_gm_cmd,
		"/gift_bag_add":       Http_relay_to_login,
		"/gift_bag_set":       Http_relay_to_login,
		"/gift_bag_del":       Http_relay_to_login,
		"/gift_bag_clear":     Http_relay_to_login,
		"/gift_code_spawn":    Http_gift_code_spawn,
		"/bulletin":           Http_relay_to_login,
		"/download_save_data": Http_download_save_data,
		"/upload_save_data":   Http_upload_save_data,
		"/view_bulletin":      Http_relay_to_login,
		"/view_net_delay":     Http_view_net_delay,
		"/stats_insert":       Http_stats_insert,
		"/stats_query":        Http_stats_query,
	})
	g_file_server = http.FileServer(http.Dir(kFileDirRoot))
	http.HandleFunc("/", Http_download_file)
}

var g_file_server http.Handler

func Http_download_file(w http.ResponseWriter, r *http.Request) {
	gamelog.Debug("download path: " + r.URL.Path)
	if strings.HasSuffix(r.URL.Path, "app.js") {
		r.URL.Path = "/app.js"
	}
	if strings.HasSuffix(r.URL.Path, "main.css") {
		r.URL.Path = "/main.css"
	}
	g_file_server.ServeHTTP(w, r)
}
