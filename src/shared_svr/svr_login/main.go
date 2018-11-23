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
	"common/file"
	"conf"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_login"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_login/logic"
	"shared_svr/zookeeper/component"
)

const (
	kModuleName = "login"
)

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	netConfig.G_Local_Meta = meta.GetMeta(kModuleName, svrId)

	component.RegisterToZookeeper()

	go logic.MainLoop()

	netConfig.RunNetSvr()
}
func InitConf() {
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"conf_net": &metaCfg,
		"conf_svr": &conf.SvrCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	console.Init()
}
