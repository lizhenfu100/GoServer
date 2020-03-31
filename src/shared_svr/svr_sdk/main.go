package main

import (
	"common"
	"common/console"
	"common/console/shutdown"
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
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"csv/conf_net.csv":   &metaCfg,
		"csv/conf_svr.csv":   &conf.SvrCsv,
		"csv/sdk/pingxx.csv": &conf2.PingxxSub,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	console.Init()
	console.RegShutdown(shutdown.Default)

	//展示重要配置数据
	buf, _ := json.MarshalIndent(&conf2.PingxxSub, "", "     ")
	fmt.Println("conf.Const: ", common.B2S(buf))
}
