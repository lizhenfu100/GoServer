/***********************************************************************
* @ web请假小工具
* @ brief
	1、浏览器 url 默认是 "GET" 方法
		· 后台收到GET请求，返回相应html文件
		· html中写</form>"POST"

	2、html中引用其它文件(如js、css、jpeg)
		· 须由http.FileServer映射到正确目录
		· 默认下载
			· http.Handle("/", http.FileServer(http.Dir(kFileDirRoot)))
		· 自定义函数做下载控制
			· g_file_server = http.FileServer(http.Dir(kFileDirRoot))
			· http.HandleFunc("/", Http_download_file)
			· func Http_download_file(w http.ResponseWriter, r *http.Request) {
			·   //...
			·   g_file_server.ServeHTTP(w, r)
			·   //...
			· }

* @ author zhoumf
* @ date 2018-11-8
***********************************************************************/
package main

import (
	"common"
	"common/file"
	"conf"
	"dbmgo"
	"flag"
	"fmt"
	"gamelog"
	"http"
	"netConfig"
	"netConfig/meta"
	"netConfig/register"
	"os"
	"strings"
	"time"
)

var (
	g_TypeLv = map[string]int{
		"日调": 1,
		"补班": 2,
		"调班": 2,
		"事假": 3,
		"特殊": 3,
		"年假": 4,
	}
)

func main() {
	meta.G_Local = &meta.Meta{
		Module:  "AFK",
		SvrName: "ChillyRoom_AFK",
	}
	ip, port, strBeginEnd := "", 0, ""
	flag.StringVar(&ip, "ip", "120.78.152.152", "ip")
	flag.IntVar(&port, "port", 7701, "port")
	flag.StringVar(&strBeginEnd, "t", "", "开始结束日期，如2018.1.15 2018.2.15")
	flag.Parse() //内部获取了所有参数：os.Args[1:]

	meta.G_Local.IP = ip
	meta.G_Local.OutIP = ip
	meta.G_Local.HttpPort = uint16(port)

	//初始化日志系统
	gamelog.InitLogger("afk")
	InitConf()
	defer time.Sleep(time.Minute)

	if strBeginEnd != "" {
		v := strings.Split(strBeginEnd, " ")
		if len(v) < 2 {
			fmt.Println("失败，日期错误，正确格式2018.1.15 2018.2.15")
		} else {
			countAll(v[0], v[1])
		}
		return
	}
	dbmgo.InitWithUser("", 27017, "other", conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	netConfig.RunNetSvr()
}
func InitConf() {
	file.G_Csv_Map = map[string]interface{}{
		"conf_svr": &conf.SvrCsv,
	}
	file.LoadOneCsv("csv/conf_svr.csv")

	register.RegHttpHandler(map[string]register.HttpHandle{
		"/afk":    Http_afk,
		"/delete": Http_del,
		"/count":  Http_count_one,
	})
	register.RegHttpRpc(map[uint16]register.HttpRpc{
		20: Http_count_all,
	})
}

// ------------------------------------------------------------
// 统计所有人的请假信息
func countAll(beginDate, endDate string) {
	addr := http.Addr(meta.G_Local.IP, meta.G_Local.HttpPort)
	http.CallRpc(addr, 20, func(buf *common.NetPack) {
		buf.WriteString(beginDate)
		buf.WriteString(endDate)
	}, func(recvBuf *common.NetPack) {
		if err := recvBuf.ReadString(); err != "ok" {
			fmt.Println(err)
		} else {
			filename := "afk_" + time.Now().Format("20060102") + ".log"
			if f, err := file.CreateFile("./", filename, os.O_WRONLY|os.O_TRUNC); err != nil {
				fmt.Println(err.Error())
				return
			} else {
				data := recvBuf.LeftBuf()
				f.Write(data)
				f.Close()
				fmt.Println("成功，当前目录可查看文本...")
				fmt.Println(string(data))
			}
		}
	})
}
