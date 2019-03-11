package main

import (
	"common"
	"common/file"
	"fmt"
	"generate_out/rpc/enum"
	"html/template"
	"http"
	"os"
	"strings"
)

var (
	g_templateData TemplateData
)

type TemplateData struct {
	LocalAddr  string
	CenterAddr string
	LoginAddrs []string
	//GameAddrs  []string
}

func getAddrs() {
	// 先从center拉取所有login地址
	http.CallRpc(kCenterAddr, enum.Rpc_meta_list, func(buf *common.NetPack) {
		buf.WriteString("login")
		buf.WriteString("")
	}, func(recvBuf *common.NetPack) {
		for cnt, i := recvBuf.ReadByte(), byte(0); i < cnt; i++ {
			recvBuf.ReadInt() //svrId
			outip := recvBuf.ReadString()
			port := recvBuf.ReadUInt16()
			recvBuf.ReadString() //name
			g_templateData.LoginAddrs = append(g_templateData.LoginAddrs, http.Addr(outip, port))
		}
	})
	// 再拉login下所有game地址
	//for _, v := range g_templateData.LoginAddrs {
	//	http.CallRpc(v, enum.Rpc_meta_list, func(buf *common.NetPack) {
	//		buf.WriteString("game")
	//		buf.WriteString("")
	//	}, func(recvBuf *common.NetPack) {
	//		for cnt, i := recvBuf.ReadByte(), byte(0); i < cnt; i++ {
	//			recvBuf.ReadInt() //svrId
	//			outip := recvBuf.ReadString()
	//			port := recvBuf.ReadUInt16()
	//			recvBuf.ReadString() //name
	//			g_templateData.GameAddrs = append(g_templateData.GameAddrs, http.Addr(outip, port))
	//		}
	//	})
	//}
}
func UpdateHtml() { //填充模板，生成可用的HTML文件，方便查看
	getAddrs()

	if names, err := file.WalkDir(kFileDirRoot+"template/", ".html"); err == nil {
		for _, name := range names {
			if t, e := template.ParseFiles(name); e != nil {
				fmt.Println("parse template error: ", e.Error())
			} else {
				name = strings.Replace(name, "template/", "", -1)
				f, _ := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
				t.Execute(f, &g_templateData)
				f.Close()
			}
		}
	}
}
