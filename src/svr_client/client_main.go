package main

import (
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
	//tcp0()
	//tcp_svr()
	fmt.Println("--- test end ---")

	MainLoop()
}
func InitConf() {
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.NilSvrCsv())
	file.LoadAllCsv()
	console.Init()
}

// ------------------------------------------------------------
func test() {
	//testRedis()
	//var v interface{}
	//var p netConfig.Rpc
	//p = nil
	//v = p
	//pp := v.(netConfig.Rpc)
	//fmt.Println(v, p, pp, pp == nil)

	//InvalidEmail()
}
func average() {
	sum, list := 0, []int{
		3979913,
	}
	for _, v := range list {
		sum += v
	}
	fmt.Println(sum/len(list))
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
	csv := email.InvalidCsv()
	file.LoadCsv("D:/soulknight/bin/UpdateCsv/csv/email/invalid.csv", &csv)
	records := make([][]string, 0, len(csv)+1)
	records = append(records, []string{"无效地址"})
	for v := range csv {
		records = append(records, []string{v})
	}
	file.UpdateCsv("D:/soulknight/bin/UpdateCsv/csv/email/", "invalid2.csv", records)
}
