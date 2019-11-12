package main

import (
	"Tools/AFK"
	"common/console"
	"common/file"
	"conf"
	"dbmgo"
	"gamelog"
	"netConfig"
	"netConfig/meta"
	"nets"
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

	meta.G_Local = meta.GetMeta(kModuleName, 0)

	dbmgo.InitWithUser("", 27017, "other", conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	afk.Init()
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

	nets.RegHttpRpc(map[uint16]nets.HttpRpc{
		116: Rpc_file_update_list, //enum.Rpc_file_update_list
		119: Rpc_file_update_list, //enum.Rpc_file_update_list
	})
	nets.RegHttpHandler(map[string]nets.HttpHandle{
		"/upload_patch_file": Http_upload_patch_file,
		"/get_login_list":    Http_get_login_list,
	})
}
