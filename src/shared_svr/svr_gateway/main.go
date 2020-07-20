package main

import (
	"common/console"
	"common/file"
	"conf"
	"flag"
	"gamelog"
	"generate_out/rpc/enum"
	_ "generate_out/rpc/shared_svr/svr_gateway"
	"netConfig"
	"netConfig/meta"
	"nets/rpc"
	"shared_svr/svr_gateway/logic"
	"shared_svr/zookeeper/component"
)

const kModuleName = "gateway"

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	if meta.G_Local = meta.GetMeta(kModuleName, svrId); meta.G_Local.HttpPort > 0 {
		rpc.G_HandleFunc[enum.Rpc_gateway_relay] = logic.RelayHttp
	} else {
		rpc.G_HandleFunc[enum.Rpc_gateway_relay] = logic.RelayTcp
	}
	component.RegisterToZookeeper()

	netConfig.RunNetSvr(true) //FIXME:考虑fasthttp
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.SvrCsv)
	file.LoadAllCsv()
	console.Init()
}
