package test

import (
	"common"
	"common/sign"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"testing"
)

// http读大数据，ReadResponse里“r.ContentLength”是-1
func Test_save_download_binary(t *testing.T) {
	for i := 0; i < 10; i++ {
		_save_download_binary(i)
	}
}
func _save_download_binary(idx int) {
	addr := "http://18.185.80.202:7090/"
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
	for i := 0; i < 10; i++ {
		_save_upload_binary(i)
	}
}
func _save_upload_binary(idx int) {
	addr := "http://18.185.80.202:7090/"
	uid := "2@qq.com"
	pf_id := "IOS"
	mac := "7F954400-D4E2-453E-9429-12E90E65B926"
	http.CallRpc(addr, enum.Rpc_save_upload_binary, func(buf *common.NetPack) {
		buf.WriteString(uid)
		buf.WriteString(pf_id)
		buf.WriteString(mac)
		s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
		buf.WriteString(sign.CalcSign(s))
		buf.WriteBuf(make([]byte, 427000))
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		fmt.Println("------------: ", idx, errCode)
	})
}
