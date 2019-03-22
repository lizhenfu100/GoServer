package main

import (
	"common/console"
	"common/file"
	"conf"
	"fmt"
	"gamelog"
	"netConfig/meta"
	"runtime/debug"
	"time"
)

const kModuleName = "client"

// go test -v ./src/svr_client/test/login_test.go
// go test -v -test.bench=".*" ./src/svr_client/test/svr_file_test.go
// 更改测试文件名的"_test"结尾，就能在main()里调测试用例了
func main() {
	gamelog.InitLogger(kModuleName)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%v: %s", r, debug.Stack())
			time.Sleep(time.Minute)
		}
	}()

	test()

	MainLoop()
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
}

// ------------------------------------------------------------
func test() {
	testQPS()
}
