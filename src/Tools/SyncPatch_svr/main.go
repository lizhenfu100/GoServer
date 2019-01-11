package main

import (
	"common/console"
	"common/file"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"netConfig/register"
)

const (
	kModuleName = "file"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger(kModuleName)
	InitConf()

	//设置本节点meta信息
	meta.G_Local = &meta.Meta{
		Module:   kModuleName,
		HttpPort: 7071,
	}
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

	register.RegHttpRpc(map[uint16]register.HttpRpc{
		enum.Rpc_file_update_list: Rpc_file_update_list,
	})
	register.RegHttpHandler(map[string]register.HttpHandle{
		"/upload_patch_file":  Http_upload_patch_file,
	})
}
