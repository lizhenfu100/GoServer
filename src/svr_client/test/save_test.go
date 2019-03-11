package test

import (
	"common"
	"common/file"
	"common/sign"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"io/ioutil"
	"os"
	//"strconv"
	"strconv"
	"strings"
	_ "svr_client/test/init"
	"testing"
)

var (
	g_roorDIr  = "../../../bin/"
	g_saveList = []string{
		"http://192.168.1.111:7090",
		"http://52.14.1.205:7090",    //1 北美
		"http://13.229.215.168:7090", //2 亚洲
		"http://54.94.211.178:7090",  //3 南美
		"http://18.185.80.202:7090",  //4 欧洲
		"http://39.96.196.250:7090",  //5 华北
		"http://47.106.35.74:7090",   //6 华南
	}
	g_saveAddr = g_saveList[0]
	g_uid      = "11"
	g_pf_id    = "TapTap"
	g_mac      = "test"
)

// go test -v ./src/svr_client/test/save_test.go
func Test_save_download_binary(t *testing.T) {
	for i := 0; i < 1; i++ {
		_save_download_binary(i)
	}
}
func _save_download_binary(idx int) {
	http.CallRpc(g_saveAddr, enum.Rpc_save_download_binary, func(buf *common.NetPack) {
		buf.WriteString(g_uid)
		buf.WriteString(g_pf_id)
		buf.WriteString(g_mac)

		s := fmt.Sprintf("uid=%s&pf_id=%s", g_uid, g_pf_id)
		buf.WriteString(sign.CalcSign(s))
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		if errCode == err.Success {
			fmt.Println("------------: ", idx, len(backBuf.LeftBuf()))
			if fi, e := file.CreateFile("D:/diner_svr", "data", os.O_TRUNC|os.O_WRONLY); e == nil {
				fi.Write(backBuf.LeftBuf())
				fi.Close()
			}
		} else {
			fmt.Println("------------: ", idx, errCode)
		}
	})
}

func tTest_save_upload_binary(t *testing.T) {
	if f, e := os.OpenFile("zipped", os.O_RDONLY, 0666); e == nil {
		if buf, e := ioutil.ReadAll(f); e == nil {
			//for i := 12096; i < 12196; i++ {
			//	uid := strconv.Itoa(i)
			fmt.Println(buf)
			return
			_save_upload_binary(g_uid, buf)
			//}
			return
		}
	}
	panic("open file err")
}
func _save_upload_binary(uid string, data []byte) {
	http.CallRpc(g_saveAddr, enum.Rpc_save_upload_binary, func(buf *common.NetPack) {
		buf.WriteString(uid)
		buf.WriteString(g_pf_id)
		buf.WriteString(g_mac)
		buf.WriteString(sign.CalcSign(fmt.Sprintf("uid=%s&pf_id=%s", uid, g_pf_id)))
		buf.WriteBuf(data)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		fmt.Println("------------: ", uid, errCode, len(data))
	})
}

// ------------------------------------------------------------
// -- 编辑/恢复
func tTest_save_set(t *testing.T) {
	if f, e := os.OpenFile("data_str.log", os.O_RDONLY, 0666); e == nil {
		if b, e := ioutil.ReadAll(f); e == nil {
			list := strings.Split(string(b), " ")
			buf := make([]byte, 0, len(b))
			for _, v := range list {
				n, _ := strconv.Atoi(v)
				buf = append(buf, byte(n))
			}
			_save_upload_binary(g_uid, buf)
			return
		}
	}
}
func tTest_restore_player_save(t *testing.T) {
	name := g_roorDIr + "player/TapTap_11/20190308_152318.save"
	if f, e := os.OpenFile(name, os.O_RDONLY, 0666); e == nil {
		if b, e := ioutil.ReadAll(f); e == nil {
			_save_upload_binary(g_uid, b)
			return
		}
	}
}
