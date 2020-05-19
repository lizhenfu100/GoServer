/***********************************************************************
* @ 聊天服
* @ brief
	· 无状态的，独立的缓存(redis)，节点彼此一致；玩家可任意选取聊天节点，方便扩展

* @ author zhoumf
* @ date 2018-3-26
***********************************************************************/
package main

import (
	"common/console"
	"common/file"
	"conf"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_chat"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_chat/logic"
	"shared_svr/zookeeper/component"
)

const kModuleName = "chat"

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(kModuleName, svrId)

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
