package test

import (
	"common"
	"fmt"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"testing"
)

func Test_update_list(t *testing.T) {
	addr := netConfig.GetHttpAddr("file", 1)
	http.CallRpc(addr, enum.Rpc_file_update_list, func(buf *common.NetPack) {
		buf.WriteString("") //version
		buf.WriteString("") //目标文件夹
	}, func(backBuf *common.NetPack) {
		cnt := backBuf.ReadUInt32()
		for i := uint32(0); i < cnt; i++ {
			name := backBuf.ReadString()
			md5 := backBuf.ReadUInt32()
			fmt.Println("Refresh file: ", name, " ", md5)
		}
	})
}
func Benchmark_update_list(b *testing.B) {
	for i := 0; i < b.N; i++ {
		go Test_update_list(nil)
	}
}

// ------------------------------------------------------------
// 压力测试，可开多个client吃满cpu
var c = make([]chan bool, 12)

func StressTest() {
	for i := 0; i < len(c); i++ {
		go func() {
			for i := 0; i < 20000; i++ {
				Test_update_list(nil)
			}
			c[i] <- true
		}()
	}
	for i := 0; i < len(c); i++ {
		<-c[i]
	}
	fmt.Println("...ok...")
}
