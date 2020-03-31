package test

import (
	"common"
	"common/file"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"io/ioutil"
	"nets/http"
	"os"
	"strconv"
	"strings"
	_ "svr_client/test/init"
	"testing"
)

var (
	g_roorDIr  = "../../../bin/"
	g_saveAddr = "http://47.106.35.74:7090"
	g_uid      = "5307437"
	g_pf_id    = "Android"
	g_mac      = "48cf955670323a684399-uc_dj"

	g_clientVersion = "0.2.2"
)

// go test -v ./src/svr_client/test/save_test.go
func Test_save_download_binary(t *testing.T) {
	http.CallRpc(g_saveAddr, enum.Rpc_save_download_binary, func(buf *common.NetPack) {
		buf.WriteString(g_uid)
		buf.WriteString(g_pf_id)
		buf.WriteString(g_mac)
		buf.WriteString("") //sign
		buf.WriteString(g_clientVersion)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		if errCode == err.Success {
			fmt.Println("------------ ok: ", len(backBuf.LeftBuf()))
			if fi, e := file.CreateFile("D:/diner_svr", g_uid+".save", os.O_TRUNC|os.O_WRONLY); e == nil {
				fi.Write(backBuf.LeftBuf())
				fi.Close()
			}
		} else {
			fmt.Println("------------ err: ", errCode)
		}
	})
}

func tTest_save_upload_binary(t *testing.T) {
	if f, e := os.Open("D:/diner_svr/" + "5307437.save"); e == nil {
		defer f.Close()
		if buf, e := ioutil.ReadAll(f); e == nil {
			//for i := 12096; i < 12196; i++ {
			//	uid := strconv.Itoa(i)
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
		buf.WriteString("") //sign
		buf.WriteString("") //extra
		buf.WriteLenBuf(data)
		buf.WriteString(g_clientVersion)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		fmt.Println("------------: ", uid, errCode, len(data))
	})
}

// ------------------------------------------------------------
// -- 编辑/恢复
func tTest_save_set(t *testing.T) {
	if f, e := os.OpenFile("data_str.log", os.O_RDONLY, 0666); e == nil {
		defer f.Close()
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
		defer f.Close()
		if b, e := ioutil.ReadAll(f); e == nil {
			_save_upload_binary(g_uid, b)
			return
		}
	}
}
