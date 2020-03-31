package main

import (
	"common"
	"common/console"
	"common/file"
	"common/tool/email"
	"conf"
	"fmt"
	"gamelog"
	"github.com/go-redis/redis"
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
	InitConf()
	meta.G_Local = meta.GetMeta("game", 1)

	fmt.Println("--- test begin ---")
	test()
	fmt.Println("--- test end ---")

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
	fmt.Println(common.CompareVersion("2.5.1", "2.6.0"))
	//testRedis()
	//var v interface{}
	//var p netConfig.Rpc
	//p = nil
	//v = p
	//pp := v.(netConfig.Rpc)
	//fmt.Println(v, p, pp, pp == nil)

	//InvalidEmail()
	//testQPS()
}

var client = redis.NewClient(&redis.Options{
	Addr:     ":6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func testRedis() {
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	client.ZAdd("test")
	if v, e := client.Get("test").Result(); e != nil {
		fmt.Println(e.Error())
	} else {
		fmt.Println(v)
	}
	client.Set("test", 1, time.Hour).Val()
}

func InvalidEmail() { //剔除重复地址
	file.LoadCsv("D:/soulknight/bin/UpdateCsv/csv/email/invalid.csv", &email.G_InvalidCsv)
	records := make([][]string, 0, len(email.G_InvalidCsv)+1)
	records = append(records, []string{"无效地址"})
	for v := range email.G_InvalidCsv {
		records = append(records, []string{v})
	}
	file.UpdateCsv("D:/soulknight/bin/UpdateCsv/csv/email/", "invalid2.csv", records)
}
