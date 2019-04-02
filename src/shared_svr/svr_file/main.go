package main

import (
	"common/console"
	"common/file"
	"conf"
	"flag"
	"gamelog"
	"generate_out/rpc/enum"
	_ "generate_out/rpc/shared_svr/svr_file"
	"netConfig"
	"netConfig/meta"
	"netConfig/register"
	"shared_svr/svr_file/logic"
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

	netConfig.RunNetSvr()
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

	register.RegHttpRpc(map[uint16]register.HttpRpc{
		enum.Rpc_file_update_list: logic.Rpc_file_update_list,
		116: logic.Rpc_file_update_list, //旧版本
	})
}
