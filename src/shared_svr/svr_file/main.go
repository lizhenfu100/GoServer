package main

import (
	"common/console"
	"common/file"
	"conf"
	"flag"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_file"
	"netConfig"
	"netConfig/meta"
	"nets"
	"shared_svr/svr_file/logic"
	"shared_svr/zookeeper/component"
)

const kModuleName = "file"

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

	netConfig.RunNetSvr(true)
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.NilSvrCsv())
	file.LoadAllCsv()
	console.Init()

	nets.RegRpc(map[uint16]nets.RpcFunc{
		116: logic.Rpc_file_update_list, //enum.Rpc_file_update_list
		119: logic.Rpc_file_update_list, //enum.Rpc_file_update_list
	})
}
