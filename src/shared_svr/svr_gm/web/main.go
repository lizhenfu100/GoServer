/***********************************************************************
* @ GM系统
* @ brief
	· 先从center拉取所有login地址，再拉login下所有game地址
	· 填充模板，生成真正的HTML文件，方便查看

* @ 内网gm工具，得是多进程的（彼此间可能rpc同名不同值）
	· 每个进程负责一个游戏，各游戏提供index.html作入口
	· 注册到 GM/index.html
	· 纯单机性的，业务统一，才可像现在这样放一个进程处理

* @ author zhoumf
* @ date 2019-2-20
***********************************************************************/
package web

import (
	"Tools/AFK/qqmsg"
	"net/http"
	"netConfig/meta"
	"nets"
	mhttp "nets/http"
	"strings"
)

const (
	kFileDirRoot = "html/GM/"
	kTemplateDir = "html/GM/template/"
)

func Init() {
	g_common.LocalAddr = mhttp.Addr(meta.G_Local.OutIP, meta.G_Local.HttpPort)
	init2()

	for k, v := range g_map { //外网节点地址
		v.TCommon = g_common
		if !v.LoadAddrs() {
			v.GetAddrs()
		}
		UpdateHtmls("game", "game."+v.GameName, &v) //game中的页面往各游戏都导一份
		UpdateHtmls("game."+v.GameName, "game."+v.GameName, &v)
		g_map[k] = v
	}
	UpdateHtmls("account/", "account/", g_common)
	UpdateHtml("index", "index", g_common)
	UpdateHtml("relay_gm_cmd", "relay_gm_cmd", g_common)
	Update_js_css("web", "web", g_common)
	go qqmsg.LoopMsg()
}
func init2() {
	nets.RegHttpHandler(map[string]nets.HttpHandle{
		"/backup_conf":    relay_to_save,
		"/backup_auto":    relay_to_save,
		"/backup_force":   relay_to_save,
		"/relay_gm_cmd":   relay_gm_cmd,
		"/gift_bag_add":   relay_to_login,
		"/gift_bag_set":   relay_to_login,
		"/gift_bag_view":  relay_to_login,
		"/gift_bag_del":   relay_to_login,
		"/gift_bag_clear": relay_to_login,
		"/bulletin":       relay_to_login,
		"/view_bulletin":  relay_to_login,
	})
	g_file_server = http.FileServer(http.Dir(kFileDirRoot))
	http.HandleFunc("/", _download_file)
}

var g_file_server http.Handler

func _download_file(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "app.js") {
		r.URL.Path = "/app.js"
	}
	if strings.HasSuffix(r.URL.Path, "stats.js") {
		r.URL.Path = "/stats.js"
	}
	if strings.HasSuffix(r.URL.Path, "main.css") {
		r.URL.Path = "/main.css"
	}
	if strings.HasSuffix(r.URL.Path, "stats.css") {
		r.URL.Path = "/stats.css"
	}
	g_file_server.ServeHTTP(w, r)
}
