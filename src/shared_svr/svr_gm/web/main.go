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
	"common/assert"
	"common/std/random"
	"fmt"
	"net/http"
	"nets"
	"strings"
)

const (
	FileDirRoot  = "html/GM/"
	kTemplateDir = "html/GM/template/"
)

func Init() {
	for k, v := range g_map {
		v.TCommon, v.GameName = &g_common, k
		if !v.LoadAddrs() {
			v.GetAddrs()
		}
		UpdateHtmls("game", "game."+v.GameName, &v) //game中的页面往各游戏都导一份
		UpdateHtmls("game."+v.GameName, "game."+v.GameName, &v)
		g_map[k] = v
	}
	UpdateHtmls("account/", "account/", &g_common)
	UpdateHtml("index", "index", &g_common)

	// web入口
	if !assert.IsDebug {
		_gate = "/" + random.String(16)
	}
	http.HandleFunc("/", _download_file)
	fmt.Println(g_common.LocalAddr() + _gate)
}
func init() {
	nets.RegHttpHandler(map[string]nets.HttpHandle{
		"/backup_conf":     relay_to_save,
		"/backup_force":    relay_to_save,
		"/relay_gm_cmd":    relay_gm_cmd,
		"/client_cmd":      relay_to,
		"/gift_bag_add":    relay_to,
		"/gift_bag_set":    relay_to,
		"/gift_bag_view":   relay_to,
		"/gift_bag_del":    relay_to,
		"/gift_bag_clear":  relay_to,
		"/bulletin":        relay_to,
		"/view_bulletin":   relay_to,
		"/find_aid_in_mac": foreach_svr,
	})
	_file = http.FileServer(http.Dir(FileDirRoot))
}

var (
	_file http.Handler
	_gate = "/"
)

func _download_file(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == _gate {
		r.URL.Path = "/"
	} else if r.URL.Path == "/" {
		return //防外网扫描
	} else if strings.HasSuffix(r.URL.Path, "app.js") {
		r.URL.Path = "/app.js"
	} else if strings.HasSuffix(r.URL.Path, "stats.js") {
		r.URL.Path = "/stats.js"
	} else if strings.HasSuffix(r.URL.Path, "main.css") {
		r.URL.Path = "/main.css"
	} else if strings.HasSuffix(r.URL.Path, "stats.css") {
		r.URL.Path = "/stats.css"
	}
	_file.ServeHTTP(w, r)
}
