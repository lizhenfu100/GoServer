package test

import (
	"common/file"
	"conf"
	"fmt"
	"gamelog"
	"netConfig"
	"netConfig/meta"
	"strings"
	"sync"
	"testing"
)

func init() {
	fmt.Println("--- unit test init ---")
	gamelog.InitLogger("test")
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"conf_net": &metaCfg,
		"conf_svr": &conf.SvrCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	netConfig.G_Local_Meta = meta.GetMeta("client", 0)
}

func Test_1(t *testing.T) {
	var v interface{} = uint32(10)
	v2 := uint32(10)
	fmt.Println(v == v2)

	var map1 sync.Map
	map1.Store("a", int(10))
	vv, _ := map1.Load("a")
	fmt.Println(vv == t)

	addr := "http://192.168.1.11:2233/"
	idx1 := strings.Index(addr, "//") + 2
	idx2 := strings.LastIndex(addr, ":")
	fmt.Println(addr[idx1:idx2])
	fmt.Println(addr[idx2+1 : len(addr)-1])
}

func Test_2(t *testing.T) {
	lst := new([]int)
	lst1 := lst
	lst2 := lst
	*lst = append(*lst, 23)
	fmt.Println(lst, lst1, lst2)

	team := *lst
	*lst = append(team, 1)
	fmt.Println(lst, lst1, lst2)
}
