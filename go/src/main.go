package main

import (
	"common"
	"fmt"
	//	"gamelog"
	//  "http"
	"strings"
	// "tcp"
)

func main() {
	fmt.Println("=>", strings.Index("(c)#蘑菇adfadsf", "(c"))

	fmt.Println(common.GetExePath())

	//	common.LoadCsv("labs\\test.csv")
	//	fmt.Println(common.G_MapCsv)
	//	fmt.Println(common.G_SliceCsv)

	//初始化日志系统
	// gamelog.InitLogger("gitSundry", true)
	// gamelog.SetLevel(0)

	// tcp.NewTcpServer(":9001", 5000)
	// http.NewHttpServer(":9002")
}
