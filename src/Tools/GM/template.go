package main

import (
	"common"
	"common/file"
	"encoding/json"
	"fmt"
	"generate_out/rpc/enum"
	"html/template"
	"io/ioutil"
	"nets/http"
	"os"
	"strings"
)

var (
	g_addrs = &TemplateData{
		CenterAddr: "http://52.14.1.205:7000",
		SdkAddrs: []string{"",
			"http://120.78.152.152:7003", //北美
			"http://120.78.152.152:7003", //亚洲
			"http://120.78.152.152:7003", //欧洲
			"http://120.78.152.152:7003", //南美
			"http://120.78.152.152:7003", //中国华南
			"http://120.78.152.152:7003", //中国华北
		},
	}
)

type TemplateData struct {
	LocalAddr  string
	CenterAddr string
	LoginAddrs []string //0位空，1起始
	SdkAddrs   []string
	GameAddrs  []string
	SaveAddrs  []string
}

func GetAddrs() {
	g_addrs.LoginAddrs = []string{""} //0位空，大区编号从1起始
	g_addrs.GameAddrs = g_addrs.GameAddrs[:0]
	g_addrs.SaveAddrs = g_addrs.SaveAddrs[:0]
	// 先从center拉取所有login地址
	http.CallRpc(g_addrs.CenterAddr, enum.Rpc_meta_list, func(buf *common.NetPack) {
		buf.WriteString("login")
		buf.WriteString("")
	}, func(recvBuf *common.NetPack) {
		cnt := recvBuf.ReadByte()
		for i := byte(0); i < cnt; i++ {
			recvBuf.ReadInt() //svrId
			outip := recvBuf.ReadString()
			port := recvBuf.ReadUInt16()
			recvBuf.ReadString() //name
			g_addrs.LoginAddrs = append(g_addrs.LoginAddrs, http.Addr(outip, port))
		}
	})
	// 再拉login下所有game
	for i := 1; i < len(g_addrs.LoginAddrs); i++ {
		http.CallRpc(g_addrs.LoginAddrs[i], enum.Rpc_meta_list, func(buf *common.NetPack) {
			buf.WriteString("game")
			buf.WriteString("")
		}, func(recvBuf *common.NetPack) {
			cnt := recvBuf.ReadByte()
			for i := byte(0); i < cnt; i++ {
				recvBuf.ReadInt() //svrId
				outip := recvBuf.ReadString()
				port := recvBuf.ReadUInt16()
				recvBuf.ReadString() //name
				g_addrs.GameAddrs = append(g_addrs.GameAddrs, http.Addr(outip, port))
			}
		})
	}
	// 再拉game下所有save
	for _, v := range g_addrs.GameAddrs {
		http.CallRpc(v, enum.Rpc_meta_list, func(buf *common.NetPack) {
			buf.WriteString("save")
			buf.WriteString("")
		}, func(recvBuf *common.NetPack) {
			cnt := recvBuf.ReadByte()
			for i := byte(0); i < cnt; i++ {
				recvBuf.ReadInt() //svrId
				outip := recvBuf.ReadString()
				port := recvBuf.ReadUInt16()
				recvBuf.ReadString() //name
				g_addrs.SaveAddrs = append(g_addrs.SaveAddrs, http.Addr(outip, port))
			}
		})
	}

	if fi, e := file.CreateFile("log/", "gm.addr", os.O_TRUNC|os.O_WRONLY); e == nil {
		buf, _ := common.T2B(g_addrs)
		fi.Write(buf)
		fi.Close()
	}
	buf, _ := json.MarshalIndent(g_addrs, "", "     ")
	fmt.Println(common.B2S(buf))
}
func LoadAddrs() bool {
	if f, e := os.Open("log/gm.addr"); e == nil {
		if buf, e := ioutil.ReadAll(f); e == nil {
			common.B2T(buf, g_addrs)
			buf, _ = json.MarshalIndent(g_addrs, "", "     ")
			fmt.Println(common.B2S(buf))
			return true
		}
	}
	return false
}

func UpdateHtml() { //填充模板，生成可用的HTML文件，方便查看
	if names, err := file.WalkDir(kFileDirRoot+"template/", ".html"); err == nil {
		for _, name := range names {
			if t, e := template.ParseFiles(name); e != nil {
				fmt.Println("parse template error: ", e.Error())
			} else {
				name = strings.Replace(name, "template/", "", -1)
				f, _ := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
				t.Execute(f, g_addrs)
				f.Close()
			}
		}
	}
}
