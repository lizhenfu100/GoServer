/***********************************************************************
* @ 好友服务器
* @ brief
	1、独立的好友服，好友关系入库(类似微信，实现社交关系不同游戏间复用)

	2、避免维护“玩家-好友服”对应关系
		a、设计为无状态的，独立的缓存(redis)，节点彼此一致；玩家可任意选取
		b、设计为http节点，同步取数据，接口更友好；让业务节点遍历“好友服集群”，缓存成功取到数据的节点信息

* @ author zhoumf
* @ date 2018-3-26
***********************************************************************/
package main

import (
	"common/console"
	"common/file"
	"conf"
	"dbmgo"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_friend"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_friend/logic"
	"shared_svr/zookeeper/component"
)

const (
	kModuleName = "friend"
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

	//设置mongodb的服务器地址
	pMeta := meta.GetMeta("db_friend", 0)
	dbmgo.InitWithUser(pMeta.IP, pMeta.Port(), pMeta.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)

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
