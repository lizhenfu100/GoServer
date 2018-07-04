package main

import (
	"flag"
	"gamelog"
	"netConfig"
	"netConfig/meta"
	"netConfig/register"
)

func init() {
	flag.StringVar(&G_OutDir, "o", "./", "输出目录：默认为exe所在目录")
}

func main() {
	//初始化日志系统
	gamelog.InitLogger("afk")
	flag.Parse()
	InitConf()

	netConfig.RunNetSvr()
}
func InitConf() {
	netConfig.G_Local_Meta = &meta.Meta{
		Module:   "AFK",
		SvrName:  "ChillyRoom_AFK",
		HttpPort: 7601,
	}
	register.RegHttpHandler(map[string]register.HttpHandle{
		"/ask_for_leave": Http_ask_for_leave,
	})
}
