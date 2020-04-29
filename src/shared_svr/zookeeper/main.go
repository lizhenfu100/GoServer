/***********************************************************************
* @ zookeeper
* @ brief
	1、每个节点都同zookeeper相连，config_net.csv仅zookeeper解析
	2、其它节点启动后，主动连接zoo，zoo做两件事情：
		、查询哪些节点要连接此新节点，并告知它们新节点的meta
		、告知新节点，待连接节点的meta

* @ 节点断开
	· 有些节点不是能随便删的，比如gateway
	· 运行时少了台，玩家哈希改变，数据会乱掉

* @ author zhoumf
* @ date 2017-11-27
***********************************************************************/
package main

import (
	"common/console"
	"common/file"
	"conf"
	"gamelog"
	"netConfig"
	"netConfig/meta"
)

const kModuleName = "zookeeper" //tcp_multi

func main() {
	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(kModuleName, 0)

	netConfig.RunNetSvr(true)
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.SvrCsv())
	file.LoadAllCsv()
	console.Init()
}
