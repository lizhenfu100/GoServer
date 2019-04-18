package init

import (
	"common/file"
	"conf"
	"fmt"
	"gamelog"
	"netConfig/meta"
	"nets/http"
	http2 "nets/http/http"
)

func init() {
	fmt.Println("--- unit test init ---")
	gamelog.InitLogger("test")
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"csv/conf_net.csv": &metaCfg,
		"csv/conf_svr.csv": &conf.SvrCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	meta.G_Local = meta.GetMeta("client", 0)
	http.InitClient(http2.Client)
}
