package test

import (
	"common"
	"common/sign"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"io/ioutil"
	"os"
	"strconv"
	_ "svr_client/test/init"
	"testing"
)

// http读大数据，ReadResponse里“r.ContentLength”是-1
func Test_save_download_binary(t *testing.T) {
	for i := 0; i < 10; i++ {
		_save_download_binary(i)
	}
}
func _save_download_binary(idx int) {
	addr := "http://18.185.80.202:7090"
	uid := "2@qq.com"
	pf_id := "IOS"
	mac := "7F954400-D4E2-453E-9429-12E90E65B926"
	http.CallRpc(addr, enum.Rpc_save_download_binary, func(buf *common.NetPack) {
		buf.WriteString(uid)
		buf.WriteString(pf_id)
		buf.WriteString(mac)

		s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
		buf.WriteString(sign.CalcSign(s))
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		if errCode == err.Success {
			fmt.Println("------------: ", idx, len(backBuf.LeftBuf()))
		} else {
			fmt.Println("------------: ", idx, errCode)
		}
	})
}

func Test_save_upload_binary(t *testing.T) {
	if f, e := os.OpenFile("zipped", os.O_RDONLY, 0666); e == nil {
		if buf, e := ioutil.ReadAll(f); e == nil {
			for i := 12096; i < 12196; i++ {
				uid := strconv.Itoa(i)
				_save_upload_binary(uid, buf)
			}
			return
		}
	}
	panic("open file err")
}
func _save_upload_binary(uid string, data []byte) {
	addr := "http://13.229.215.168:7090"
	pf_id := "Android"
	mac := uid
	http.CallRpc(addr, enum.Rpc_save_upload_binary, func(buf *common.NetPack) {
		buf.WriteString(uid)
		buf.WriteString(pf_id)
		buf.WriteString(mac)
		buf.WriteString(sign.CalcSign(fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)))
		buf.WriteBuf(data)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		fmt.Println("------------: ", uid, errCode)
	})
}
