/***********************************************************************
* @ Nginx
* @ /etc/nginx/sites-enabled
server {
    listen 80 default_server;
    listen [::]:80 default_server ipv6only=on;
    index index.html index.htm;
    server_name www.chillyroom.com;
    error_page 404 502 /404.html;
    root /home/ubuntu/chillyroom-gate-v2/public/;
    location / {
        root /home/ubuntu/web/public/;
    }
    location /api/ {
        proxy_pass http://127.0.0.1:7071/;
    }
    location = /404.html {
        root /home/ubuntu/web/public/;
    }
}
server {
    listen 80;
    server_name afk.chillyroom.com;
    location / {
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $host;
        proxy_pass http://161.189.123.149:7708;
    }
}
* @ author zhoumf
* @ date 2020-4-2
***********************************************************************/
package main

import (
	"encoding/json"
	"net/http"
)

// ------------------------------------------------------------
// -- 登录服列表 http://chillyroom.com/api/
type TLogins struct {
	GameName string
	Logins   map[string][]string //<大区名, 地址>
}

var G_Logins map[string]*TLogins = nil //<游戏名, 登录列表>

func Http_get_login_list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v, ok := G_Logins[q.Get("name")]; ok {
		str, _ := json.MarshalIndent(v.Logins, "", "     ")
		w.Write(str)
	}
}
