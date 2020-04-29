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
	_ "generate_out/rpc/svr_game"
	"netConfig"
	"netConfig/meta"
	"shared_svr/zookeeper/component"
	conf2 "svr_game/conf"
	"svr_game/logic"
)

const kModuleName = "game"

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

	netConfig.RunNetSvr(false)
	logic.MainLoop()
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.SvrCsv())
	file.RegCsvType("csv/game/const.csv", conf2.Csv())
	file.LoadAllCsv()
	console.Init()
	console.RegShutdown(logic.Shutdown)

	if list := meta.GetMetas(conf.GameName, ""); len(list) > 0 {
		conf2.Csv().LoginSvrId = list[0].SvrID % common.KIdMod
	} else {
		panic("LoginSvrId nil")
	}
	//展示重要配置数据
	buf, _ := json.MarshalIndent(conf2.Csv(), "", "     ")
	fmt.Println("conf.Const: ", common.B2S(buf))
}
