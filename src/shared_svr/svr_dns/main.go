package main

import (
	"common"
	"common/console"
	"common/console/shutdown"
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
	file.G_Csv_Map = map[string]interface{}{
		"csv/outip.csv": &logic.G_Logins,
	}
	file.LoadAllCsv()
	console.Init()
	console.RegShutdown(shutdown.Default)

	//展示重要配置数据
	b, _ := json.MarshalIndent(&logic.G_Logins, "", "     ")
	fmt.Println("G_Logins: ", common.B2S(b))
}
