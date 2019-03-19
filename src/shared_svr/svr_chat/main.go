/***********************************************************************
* @ 聊天服
* @ brief
	1、设计为无状态的，独立的缓存(redis)，节点彼此一致；玩家可任意选取聊天节点，方便扩展

	2、可gateway转发，也可客户端直连svr_chat(hash取模+chat互连)；若直连须补充登录流程

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

	go netConfig.RunNetSvr()
	logic.MainLoop()
}
func InitConf() {
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"csv/conf_net.csv": &metaCfg,
		"csv/conf_svr.csv": &conf.SvrCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	console.Init()
}
