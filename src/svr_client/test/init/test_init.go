package init

import (
	"common/file"
	"conf"
	"fmt"
	"gamelog"
	"netConfig/meta"
)

func init() {
	fmt.Println("--- unit test init ---")
	gamelog.InitLogger("test")
	var metaCfg meta.Metas
	file.RegCsvType("csv/conf_net.csv", metaCfg)
	file.RegCsvType("csv/conf_svr.csv", conf.SvrCsv())
	file.LoadAllCsv()
	meta.G_Local = meta.GetMeta("client", 0)
}
