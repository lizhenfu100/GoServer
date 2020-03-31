/***********************************************************************
* @ 登录服
* @ brief
	1、作为账号服务器(svr_center)的代理，转发各svr_game请求到svr_center

	2、管理游戏服列表 svr_game list，提供给客户端选择

	3、登录排队逻辑

* @ author zhoumf
* @ date 2017-10-23
***********************************************************************/
package main

import (
	"common/console"
	"common/console/shutdown"
	"common/file"
	"common/tool/email"
	"conf"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_login"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_login/logic"
	"shared_svr/zookeeper/component"
)

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger("login")
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(conf.GameName, svrId)

	component.RegisterToZookeeper()

	netConfig.RunNetSvr(false)
	logic.MainLoop()
}
func InitConf() {
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"csv/conf_net.csv":      &metaCfg,
		"csv/conf_svr.csv":      &conf.SvrCsv,
		"csv/email/email.csv":   &email.G_EmailCsv,
		"csv/email/invalid.csv": &email.G_InvalidCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	console.Init()
	console.RegShutdown(shutdown.Default)
}
