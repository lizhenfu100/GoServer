package main

import (
	"common"
	"common/console"
	"common/file"
	"common/tool/email"
	"conf"
	"encoding/json"
	"flag"
	"fmt"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_save"
	"netConfig"
	"netConfig/meta"
	conf2 "shared_svr/svr_save/conf"
	"shared_svr/svr_save/logic"
	"shared_svr/zookeeper/component"
)

const kModuleName = "save"

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = meta.GetMeta(kModuleName, svrId)
	if meta.G_Local.HttpPort != meta.KSavePort {
		panic("svr_save port err")
	}
	component.RegisterToZookeeper()

	netConfig.RunNetSvr(false)
	logic.MainLoop()
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.NilSvrCsv())
	file.RegCsvType("csv/save/const.csv", conf2.NilCsv())
	file.RegCsvType("csv/email/email.csv", email.NilEmailCsv())
	file.RegCsvType("csv/email/invalid.csv", email.NilInvalidCsv())
	file.LoadAllCsv()
	console.Init()

	//展示重要配置数据
	buf, _ := json.MarshalIndent(conf2.Csv(), "", "     ")
	fmt.Println("conf.Const: ", common.B2S(buf))
}
