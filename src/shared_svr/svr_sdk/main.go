package main

import (
	"common"
	"common/console"
	"common/file"
	"conf"
	"encoding/json"
	"flag"
	"fmt"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_sdk"
	"netConfig"
	"netConfig/meta"
	conf2 "shared_svr/svr_sdk/conf"
	"shared_svr/svr_sdk/logic"
	"shared_svr/svr_sdk/platform"
	"shared_svr/zookeeper/component"
)

const kModuleName = "sdk"

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 0, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(kModuleName, svrId)

	component.RegisterToZookeeper()

	netConfig.RunNetSvr(false)
	platform.Init()
	logic.MainLoop()
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.SvrCsv())
	file.RegCsvType("csv/sdk/pingxx.csv", conf2.PingxxSub())
	file.LoadAllCsv()
	console.Init()

	//展示重要配置数据
	buf, _ := json.MarshalIndent(conf2.PingxxSub(), "", "     ")
	fmt.Println("conf.Const: ", common.B2S(buf))
}
