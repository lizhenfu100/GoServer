package main

import (
	"common"
	"common/console"
	"common/file"
	"encoding/json"
	"fmt"
	"gamelog"
	_ "generate_out/rpc/shared_svr/svr_dns"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_dns/logic"
)

func main() {
	gamelog.InitLogger("dns")
	InitConf()
	meta.G_Local = &meta.Meta{HttpPort: 7233}
	netConfig.RunNetSvr(true)
}
func InitConf() {
	file.RegCsvType("csv/outip.csv", logic.Logins())
	file.LoadAllCsv()
	console.Init()

	//展示重要配置数据
	b, _ := json.MarshalIndent(logic.Logins(), "", "     ")
	fmt.Println("G_Logins: ", common.B2S(b))
}
