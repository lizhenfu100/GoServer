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
	"path/filepath"
	"strings"
)

var (
	g_common = _TCommon{
		CenterAddr: "http://52.14.1.205:7000",
		SdkAddrs: []string{"",
			"http://120.78.152.152:7003", //1 北美
			"http://120.78.152.152:7003", //2 亚洲
			"http://120.78.152.152:7003", //3 欧洲
			"http://120.78.152.152:7003", //4 南美
			"http://120.78.152.152:7003", //5 中国华北
			"http://120.78.152.152:7003", //6 中国华南
		},
	}
	g_list = []TemplateData{
		{
			GameName: "HappyDiner", //游戏名，以及对应的大区
			LoginAddrs: []string{"", //0位空，大区编号从1起始
				"http://52.14.1.205:7030",    //1 北美
				"http://13.229.215.168:7030", //2 亚洲
				"http://18.185.80.202:7030",  //3 欧洲
				"http://54.94.211.178:7030",  //4 南美
				"http://39.96.196.250:7030",  //5 中国华北
				"http://47.106.35.74:7030",   //6 中国华南
			},
		},
		{
			GameName: "SoulKnight",
			LoginAddrs: []string{"",
				"", //1 北美
				"", //2 亚洲
				"", //3 欧洲
				"", //4 南美
				"http://39.97.111.110:7030", //5 中国华北
				"http://39.108.87.225:7030", //6 中国华南
			},
		},
	}
)

type _TCommon struct {
	LocalAddr  string
	CenterAddr string
	SdkAddrs   []string //支付sdk暂时各项目共用
}
type TemplateData struct {
	_TCommon
	GameName   string
	LoginAddrs []string //0位空，1起始
	GameAddrs  []string
	SaveAddrs  []string
}

func (self *TemplateData) GetAddrs() {
	self.GameAddrs = self.GameAddrs[:0]
	self.SaveAddrs = self.SaveAddrs[:0]
	// 先拉login下所有game
	for i := 1; i < len(self.LoginAddrs); i++ {
		if self.LoginAddrs[i] != "" {
			http.CallRpc(self.LoginAddrs[i], enum.Rpc_meta_list, func(buf *common.NetPack) {
				buf.WriteString("game")
				buf.WriteString("")
			}, func(recvBuf *common.NetPack) {
				cnt := recvBuf.ReadByte()
				for i := byte(0); i < cnt; i++ {
					recvBuf.ReadInt() //svrId
					outip := recvBuf.ReadString()
					port := recvBuf.ReadUInt16()
					recvBuf.ReadString() //name
					self.GameAddrs = append(self.GameAddrs, http.Addr(outip, port))
				}
			})
		}
	}
	// 再拉game下所有save
	for _, v := range self.GameAddrs {
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
				self.SaveAddrs = append(self.SaveAddrs, http.Addr(outip, port))
			}
		})
	}
	// 拉到的数据写本地文件
	if fi, e := file.CreateFile("log/", self.GameName+".addr", os.O_TRUNC|os.O_WRONLY); e == nil {
		buf, _ := common.T2B(self)
		fi.Write(buf)
		fi.Close()
	}
	buf, _ := json.MarshalIndent(self, "", "     ")
	fmt.Println(common.B2S(buf))
}
func (self *TemplateData) LoadAddrs() bool {
	if f, e := os.Open("log/" + self.GameName + ".addr"); e == nil {
		if buf, e := ioutil.ReadAll(f); e == nil {
			common.B2T(buf, self)
			buf, _ = json.MarshalIndent(self, "", "     ")
			fmt.Println(common.B2S(buf))
			return true
		}
	}
	return false
}

func UpdateHtmls(dirIn, dirOut string, ptr interface{}) { //填充模板，生成可用的HTML文件，方便查看
	if names, err := file.WalkDir(kFileDirRoot+"template/"+dirIn, ".html"); err == nil {
		for _, name := range names {
			if t, e := template.ParseFiles(name); e != nil {
				fmt.Println("parse template error: ", e.Error())
			} else {
				fullname := strings.Replace(name, "template/"+dirIn, dirOut, -1)
				dir, name := filepath.Split(fullname)
				f, _ := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC)
				t.Execute(f, ptr)
				f.Close()
			}
		}
	}
}
func UpdateHtml(fileIn, fileOut string, ptr interface{}) {
	fullname := kFileDirRoot + "template/" + fileIn + ".html"
	if t, e := template.ParseFiles(fullname); e != nil {
		fmt.Println("parse template error: ", e.Error())
	} else {
		fullname = strings.Replace(fullname, "template/"+fileIn, fileOut, -1)
		dir, name := filepath.Split(fullname)
		f, _ := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC)
		t.Execute(f, ptr)
		f.Close()
	}
}
