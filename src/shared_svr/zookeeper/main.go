/***********************************************************************
* @ zookeeper
* @ brief
	1、每个节点都同zookeeper相连，config_net.csv仅zookeeper解析

	2、其它节点启动后，主动连接zoo，zoo做两件事情：

		、查询哪些节点要连接此新节点，并告知它们新节点的meta

		、告知新节点，待连接节点的meta

	3、子节点缓存zookeeper下发的meta

* @ author zhoumf
* @ date 2017-11-27
***********************************************************************/
package main

import (
	"common/console"
	"common/file"
	"conf"
	"gamelog"
	_ "generate_out/rpc/shared_svr/zookeeper"
	"netConfig"
	"netConfig/meta"
	"shared_svr/zookeeper/logic"
)

const kModuleName = "zookeeper"

func main() {
	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(kModuleName, 0)

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
