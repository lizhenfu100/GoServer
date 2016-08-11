/***********************************************************************
* @ 多进程服务器架构
* @ brief
	1、主逻辑游戏服使用http同Client通信

	2、服务器进程间用tcp通信

	3、未来扩展：Battle设计为多个，初始化完毕后http.Post自己的信息到Gamesvr（甚至能临时加机器）

	TODO：
		区分inner_ip/outer_ip

* @ author zhoumf
* @ date 2016-8-11
***********************************************************************/
package main

import (
	"encoding/json"
)

type TAddrInfo struct {
	IP       string `json:"ip"`
	TcpPort  int    `json:"tcpPort"`
	HttpPort int    `json:"httpPort"`
	Maxconn  int    `json:"maxconn"`
}
type TSvrNetCfg struct {
	Listen  map[string]TAddrInfo `json:"listen"`
	Connect map[string]TAddrInfo `json:"connect"`
}

var g_svrModule = `{
	"Account": {
		"listen": {
			"Gamesvr": {"ip":"localhost", "tcpPort":8081}
		},
		"connect": {

		}
	},
	"Gamesvr": {
		"listen": {
			"Client": {"ip":"localhost", "httpPort":8081, "maxconn":5000}
		},
		"connect": {
			"Chat": {"ip":"localhost", "tcpPort":8081},
			"Battle": {"ip":"localhost", "tcpPort":8081}
		}
	},
	"Chat": {
		"listen": {
			"Client": {"ip":"localhost", "tcpPort":8081, "maxconn":5000},
			"Gamesvr": {"ip":"localhost", "tcpPort":8081}
		},
		"connect": {

		}
	},
	"Battle": {
		"listen": {
			"Client": {"ip":"localhost", "tcpPort":8081, "maxconn":5000},
			"Gamesvr": {"ip":"localhost", "tcpPort":8081}
		},
		"connect": {
			
		}
	},
	"Cross": {

	},
	"Client": {
		"listen": {

		},
		"connect": {
			"Chat": {"ip":"localhost", "tcpPort":8081},
			"Battle": {"ip":"localhost", "tcpPort":8081},
			"Gamesvr": {"ip":"localhost", "httpPort":8081}
		}
	}
}`
var G_SvrNetCfg map[string]TSvrNetCfg

func ParseSvrConfig() {
	if err := json.Unmarshal([]byte(g_svrModule), &G_SvrNetCfg); err != nil {
		panic("ParseSvrConfig fail :" + err.Error())
	}
}
