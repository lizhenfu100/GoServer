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
	"common"
	"common/net/meta"
	"conf"
	"gamelog"
	_ "generate_out/rpc/zookeeper"
	"netConfig"
	"zookeeper/logic"
)

const (
	K_Module_Name  = "zookeeper"
	K_Module_SvrID = 0
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(K_Module_Name)
	if conf.IsDebug {
		gamelog.SetLevel(gamelog.Lv_Debug)
	} else {
		gamelog.SetLevel(gamelog.Lv_Info)
	}
	InitConf()

	go logic.MainLoop()

	print("----zookeeper Start-----")
	if !netConfig.CreateNetSvr(K_Module_Name, K_Module_SvrID) {
		print("----zookeeper Failed-----")
	}
}
func InitConf() {
	var metaCfg []meta.Meta
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &metaCfg,
		"conf_svr": &conf.SvrCsv,
	}
	common.LoadAllCsv()
	meta.InitConf(metaCfg)

	netConfig.G_Local_Meta = meta.GetMeta(K_Module_Name, K_Module_SvrID)
}
