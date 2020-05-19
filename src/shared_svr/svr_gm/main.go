/***********************************************************************
* @ GM
* @ brief
	1、管理单个项目的某地区，比如国内（包含华南、华北两大区）
	2、地区共享的数据，如礼包码……防跨大区二次领取

* @ author zhoumf
* @ date 2018-12-12
***********************************************************************/
package main

import (
	"common/console"
	"common/file"
	"conf"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_gm"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_gm/logic"
	"shared_svr/svr_gm/web"
	"shared_svr/zookeeper/component"
)

const kModuleName = "gm"

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(kModuleName, svrId)

	if file.IsExist(web.FileDirRoot) {
		web.Init() //GM页面
	}
	component.RegisterToZookeeper()

	netConfig.RunNetSvr(false)
	logic.MainLoop()
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.NilSvrCsv())
	file.LoadAllCsv()
	console.Init()
}
