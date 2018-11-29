package main

import (
	"common/console"
	"common/file"
	"conf"
	"fmt"
	"gamelog"
	"netConfig"
	"netConfig/meta"
	"runtime/debug"
	"time"
)

const (
	kModuleName = "client"
)

// go test -v ./src/svr_client/unit_test/http_test.go
// 更改测试文件名的"_test"结尾，就能在main()里调测试用例了
func main() {
	gamelog.InitLogger(kModuleName)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%v: %s", r, debug.Stack())
			time.Sleep(time.Minute)
		}
	}()
	InitConf()
	netConfig.G_Local_Meta = meta.GetMeta(kModuleName, 0)

	//netConfig.RunNetSvr()
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
}

// ------------------------------------------------------------
