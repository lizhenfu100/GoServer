package main

import (
	"common/console"
	"common/file"
	"conf"
	"gamelog"
	"netConfig"
	"netConfig/meta"
	"netConfig/register"
)

const kModuleName = "file"

/*
	官网执行程序路径：~/web/public/SyncPatch_svr
	官网推广资源路径：~/web/public/game/promotion/
*/

// nohup ./SyncPatch_svr > nohup.out 2>&1 &
func main() {
	gamelog.InitLogger(kModuleName)
	InitConf()

	meta.G_Local = &meta.Meta{
		Module:   kModuleName,
		HttpPort: 7071,
	}
	netConfig.RunNetSvr()
}
func InitConf() {
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"csv/conf_net.csv": &metaCfg,
		"csv/conf_svr.csv": &conf.SvrCsv,
		"csv/logins.csv":   &G_Logins,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	console.Init()

	register.RegHttpRpc(map[uint16]register.HttpRpc{
		116: Rpc_file_update_list, //enum.Rpc_file_update_list
		119: Rpc_file_update_list, //enum.Rpc_file_update_list
	})
	register.RegHttpHandler(map[string]register.HttpHandle{
		"/upload_patch_file": Http_upload_patch_file,
		"/get_login_list":    Http_get_login_list,
	})
}
