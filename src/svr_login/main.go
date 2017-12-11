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
	"common"
	"common/net/meta"
	"conf"
	"gamelog"
	_ "generate_out/rpc/svr_login"
	"netConfig"
	"zookeeper/component"
)

const (
	K_Module_Name  = "login"
	K_Module_SvrID = 0
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(K_Module_Name)
	gamelog.SetLevel(gamelog.Lv_Debug)
	InitConf()

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()

	component.RegisterToZookeeper()

	print("----Login Server Start-----")
	if !netConfig.CreateNetSvr(K_Module_Name, K_Module_SvrID) {
		print("----Login NetSvr Failed-----")
	}
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &meta.G_SvrNets,
		"conf_svr": &conf.SvrCsv,
	}
	common.LoadAllCsv()

	netConfig.G_Local_Meta = meta.GetMeta(K_Module_Name, K_Module_SvrID)
}
